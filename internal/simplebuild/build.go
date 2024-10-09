package simplebuild

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/hairyhenderson/go-codeowners"
	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/internal/files"
)

// Build compiles the versioned resources in a project configuration based on
// simplified versioning rules, after the start date.
func Build(
	ctx context.Context,
	project *config.Project,
	startDate vervet.Version,
	versioningUrl string,
	appendOutputFiles bool,
) error {
	if time.Now().Before(startDate.Date) {
		return nil
	}

	latestVersion, err := fetchLatestVersion(versioningUrl)
	if err != nil {
		return err
	}

	for _, apiConfig := range project.APIs {
		if apiConfig.Output == nil {
			fmt.Printf("No output specified for %s, skipping\n", apiConfig.Name)
			continue
		}

		for _, resource := range apiConfig.Resources {
			paths, err := ResourceSpecFiles(resource)
			if err != nil {
				return err
			}

			if err := CheckSingleVersionResourceToBeBeforeLatestVersion(paths, latestVersion); err != nil {
				return err
			}
		}

		operations, err := LoadPaths(ctx, apiConfig)
		if err != nil {
			return err
		}
		for _, op := range operations {
			op.Annotate()
		}
		docs := operations.Build(startDate)
		writer, err := NewWriter(*apiConfig.Output, appendOutputFiles)
		if err != nil {
			return err
		}

		sortDocsByVersionDate(docs)

		err = CheckBreakingChanges(docs)
		if err != nil {
			return err
		}

		// Process each document
		for _, doc := range docs {
			err := doc.ApplyOverlays(ctx, apiConfig.Overlays)
			if err != nil {
				return err
			}

			if doc.Doc.Extensions == nil {
				doc.Doc.Extensions = make(map[string]interface{})
			}
			doc.Doc.Extensions[vervet.ExtSnykApiVersion] = doc.VersionDate.Format(time.DateOnly)

			refResolver := NewRefResolver()
			err = refResolver.ResolveRefs(doc.Doc)
			if err != nil {
				return err
			}

			err = writer.Write(doc)
			if err != nil {
				return err
			}
		}

		err = writer.Finalize()
		if err != nil {
			return err
		}
	}
	return nil
}

func sortDocsByVersionDate(docs DocSet) {
	slices.SortFunc(docs, func(a, b VersionedDoc) int {
		if a.VersionDate.Before(b.VersionDate) {
			return -1
		}
		if a.VersionDate.After(b.VersionDate) {
			return 1
		}
		return 0
	})
}

type OpKey struct {
	Path   string
	Method string
}

type VersionedOp struct {
	Version      vervet.Version
	Operation    *openapi3.Operation
	ResourceName string
}

type VersionSet []VersionedOp

type Operations map[OpKey]VersionSet

type VersionedDoc struct {
	VersionDate time.Time
	Doc         *openapi3.T
}
type DocSet []VersionedDoc

func (ops Operations) Build(startVersion vervet.Version) DocSet {
	filteredOps := filterBetaAndGAVersions(ops)
	versionDates := filteredOps.VersionDates()
	versionDates = filterVersionByStartDate(versionDates, startVersion.Date)
	output := make(DocSet, len(versionDates))
	for idx, versionDate := range versionDates {
		output[idx] = VersionedDoc{
			Doc:         &openapi3.T{},
			VersionDate: versionDate,
		}
		for path, spec := range filteredOps {
			op := spec.GetLatest(versionDate)
			if op == nil {
				continue
			}
			output[idx].Doc.AddOperation(path.Path, path.Method, op)
		}
	}
	return output
}

func filterBetaAndGAVersions(ops Operations) Operations {
	filteredOps := make(Operations, len(ops))
	for opKey, versionSet := range ops {
		filteredVersionSet := VersionSet{}
		for _, versionedOp := range versionSet {
			if versionedOp.Version.Stability != vervet.StabilityGA &&
				versionedOp.Version.Stability != vervet.StabilityBeta {
				continue
			}
			filteredVersionSet = append(filteredVersionSet, versionedOp)
		}
		if len(filteredVersionSet) != 0 {
			filteredOps[opKey] = filteredVersionSet
		}
	}
	return filteredOps
}

func filterVersionByStartDate(dates []time.Time, startDate time.Time) []time.Time {
	resultDates := []time.Time{startDate}
	for _, d := range dates {
		if d.After(startDate) {
			resultDates = append(resultDates, d)
		}
	}
	return resultDates
}

func (ops Operations) VersionDates() []time.Time {
	versionSet := map[time.Time]struct{}{}
	for _, opSet := range ops {
		for _, op := range opSet {
			versionSet[op.Version.Date] = struct{}{}
		}
	}
	uniqueVersions := make([]time.Time, len(versionSet))
	idx := 0
	for version := range versionSet {
		uniqueVersions[idx] = version
		idx++
	}
	return uniqueVersions
}

func LoadPaths(ctx context.Context, api *config.API) (Operations, error) {
	operations := map[OpKey]VersionSet{}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	ownerFinder, err := codeowners.FromFile(cwd)
	if err != nil {
		return nil, err
	}

	for _, resource := range api.Resources {
		paths, err := ResourceSpecFiles(resource)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			versionDir := filepath.Dir(path)
			versionStr := filepath.Base(versionDir)
			resourceName := filepath.Base(filepath.Dir(filepath.Dir(path)))

			doc, err := vervet.NewDocumentFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to load spec from %q: %w", path, err)
			}

			stabilityStr, err := vervet.ExtensionString(doc.T.Extensions, vervet.ExtSnykApiStability)
			if err != nil {
				return nil, err
			}
			if stabilityStr != "ga" {
				versionStr = fmt.Sprintf("%s~%s", versionStr, stabilityStr)
			}
			version, err := vervet.ParseVersion(versionStr)
			if err != nil {
				return nil, fmt.Errorf("invalid version %q", versionStr)
			}

			doc.InternalizeRefs(ctx, vervet.ResolveRefsWithoutSourceName)
			err = doc.ResolveRefs()
			if err != nil {
				return nil, fmt.Errorf("failed to localize refs: %w", err)
			}

			for _, pathName := range doc.T.Paths.InMatchingOrder() {
				pathDef := doc.T.Paths.Value(pathName)
				for opName, opDef := range pathDef.Operations() {
					if opDef.Extensions == nil {
						opDef.Extensions = make(map[string]interface{})
					}
					opDef.Extensions[vervet.ExtSnykApiOwner] = ownerFinder.Owners(path)
					k := OpKey{
						Path:   pathName,
						Method: opName,
					}
					if operations[k] == nil {
						operations[k] = []VersionedOp{}
					}
					operations[k] = append(operations[k], VersionedOp{
						Version:      version,
						Operation:    opDef,
						ResourceName: resourceName,
					})
				}
			}
		}
	}

	return operations, nil
}

func ResourceSpecFiles(resource *config.ResourceSet) ([]string, error) {
	return files.LocalFSSource{}.Match(resource)
}

func (vs VersionSet) GetLatest(before time.Time) *openapi3.Operation {
	var latest *VersionedOp
	for _, versionedOp := range vs {
		if versionedOp.Version.Date.After(before) {
			continue
		}
		if latest == nil {
			latest = &versionedOp
			continue
		}
		// Higher stabilities always take precedent
		if versionedOp.Version.Stability.Compare(latest.Version.Stability) < 0 {
			continue
		}
		if versionedOp.Version.Compare(latest.Version) > 0 {
			latest = &versionedOp
		}
	}
	if latest == nil {
		return nil
	}
	return latest.Operation
}

// Annotate adds Snyk specific extensions to openapi operations. These
// extensions are:
//   - x-snyk-api-version: version where the operation was defined
//   - x-snyk-api-releases: all versions of this api
//   - x-snyk-deprecated-by: if there is a later version of this operation, the
//     version of that operation
//   - x-snyk-sunset-eligible: the date after this operation can be sunset
//   - x-snyk-api-resource: what resource this operation acts on
//   - x-snyk-api-lifecycle: status of the operation, can be one of:
//     [ unreleased, released, deprecated, sunset ]
func (vs VersionSet) Annotate() {
	slices.SortFunc(vs, func(a, b VersionedOp) int {
		return a.Version.Compare(b.Version)
	})

	count := len(vs)

	releases := make([]string, count)
	for idx, op := range vs {
		releases[idx] = op.Version.String()
	}

	for idx, op := range vs {
		if op.Operation.Extensions == nil {
			op.Operation.Extensions = make(map[string]interface{}, 8)
		}
		op.Operation.Extensions[vervet.ExtSnykApiResource] = op.ResourceName
		op.Operation.Extensions[vervet.ExtSnykApiVersion] = op.Version.String()
		op.Operation.Extensions[vervet.ExtSnykApiReleases] = releases
		op.Operation.Extensions[vervet.ExtSnykApiLifecycle] = op.Version.LifecycleAt(time.Time{}).String()
		op.Operation.Extensions[vervet.ExtApiStabilityLevel] = MapStabilityLevel(op.Version.Stability)
		op.Operation.Extensions[vervet.ExtSnykApiStability] = op.Version.Stability.String()

		if idx < (count - 1) {
			laterVersion := vs[idx+1].Version
			// Sanity check the later version actually deprecates this one
			if !op.Version.DeprecatedBy(laterVersion) {
				continue
			}
			op.Operation.Extensions[vervet.ExtSnykDeprecatedBy] = laterVersion.String()
			sunsetDate, ok := op.Version.Sunset(laterVersion)
			if ok {
				op.Operation.Extensions[vervet.ExtSnykSunsetEligible] = sunsetDate.Format("2006-01-02")
			}
		}
	}
}

// MapStabilityLevel maps the vervet stability level to the x-API stability level header.
func MapStabilityLevel(s vervet.Stability) string {
	switch s {
	case vervet.StabilityGA:
		return "stable"
	case vervet.StabilityBeta:
		return "beta"
	default:
		return ""
	}
}

func CheckBreakingChanges(docs DocSet) error {
	for i := 1; i < len(docs); i++ {
		prevDoc := docs[i-1]
		currDoc := docs[i]

		s1 := &load.SpecInfo{Spec: prevDoc.Doc}
		s2 := &load.SpecInfo{Spec: currDoc.Doc}

		diffReport, sourcesMap, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
		if err != nil {
			return err
		}
		changes := checker.CheckBackwardCompatibilityUntilLevel(
			checker.GetDefaultChecks(), diffReport, sourcesMap, checker.INFO)
		breakingChange := false
		for _, change := range changes {
			if change.IsBreaking() {
				breakingChange = true
			}
		}
		if !breakingChange {
			return fmt.Errorf("no breaking change detected between versions %s and %s: \n %s",
				prevDoc.VersionDate, currDoc.VersionDate, changes)
		}
	}
	return nil
}

func fetchLatestVersion(versioningURL string) (vervet.Version, error) {
	resp, err := http.Get(versioningURL)
	if err != nil {
		return vervet.Version{}, fmt.Errorf("failed to fetch versioning information from %q: %w", versioningURL, err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return vervet.Version{}, fmt.Errorf("failed to fetch versioning information, status code: %d", resp.StatusCode)
	}

	var versions []string
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return vervet.Version{}, fmt.Errorf("failed to parse versioning information: %w", err)
	}

	var dates = make([]string, 0, len(versions))

	for _, version := range versions {
		parts := strings.Split(version, "~")
		dates = append(dates, parts[0])
	}
	sort.Strings(dates)

	latestVersion, err := vervet.ParseVersion(dates[len(dates)-1])
	if err != nil {
		return vervet.Version{}, fmt.Errorf("failed to parse latest version date %q: %w", dates[len(dates)-1], err)
	}

	return latestVersion, nil
}

func CheckSingleVersionResourceToBeBeforeLatestVersion(paths []string, latestVersion vervet.Version) error {
	resourceVersions := make(map[string][]string)

	for _, path := range paths {
		resourceDir := filepath.Dir(filepath.Dir(path))
		versionDir := filepath.Base(filepath.Dir(path))
		resourceVersions[resourceDir] = append(resourceVersions[resourceDir], versionDir)
	}

	for _, versions := range resourceVersions {
		if len(versions) == 1 {
			versionStr := versions[0]
			version, err := vervet.ParseVersion(versionStr)
			if err != nil {
				return fmt.Errorf("invalid version %q", versionStr)
			}

			if version.Date.After(latestVersion.Date) {
				return fmt.Errorf(
					"version %s is after the last released version of the global API %s. "+
						"Please change the version date to be before %s or at the same date",
					version.Date.Format("2006-01-02"),
					latestVersion.Date.Format("2006-01-02"),
					latestVersion.Date.Format("2006-01-02"),
				)
			}
		}
	}
	return nil
}
