package simplebuild

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/config"
	"github.com/snyk/vervet/v7/internal/files"
)

// Build compiles the versioned resources in a project configuration based on
// simplified versioning rules, after the start date.
func Build(ctx context.Context, project *config.Project, startDate vervet.Version, appendOutputFiles bool) error {
	if time.Now().Before(startDate.Date) {
		return nil
	}
	for _, apiConfig := range project.APIs {
		operations, err := LoadPaths(ctx, apiConfig)
		if err != nil {
			return err
		}
		for _, op := range operations {
			op.Annotate()
		}

		docs, err := operations.Build(startDate)
		if err != nil {
			return err
		}

		err = docs.ApplyOverlays(ctx, apiConfig.Overlays)
		if err != nil {
			return err
		}

		if apiConfig.Output != nil {
			err = docs.WriteOutputs(*apiConfig.Output, appendOutputFiles)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

func (ops Operations) Build(startVersion vervet.Version) (DocSet, error) {
	versionDates := ops.VersionDates()
	versionDates = filterVersionByStartDate(versionDates, startVersion.Date)
	output := make(DocSet, len(versionDates))
	for idx, versionDate := range versionDates {
		output[idx] = VersionedDoc{
			Doc:         &openapi3.T{},
			VersionDate: versionDate,
		}
		refResolver := NewRefResolver(output[idx].Doc)
		for path, spec := range ops {
			op := spec.GetLatest(versionDate)
			if op == nil {
				continue
			}
			output[idx].Doc.AddOperation(path.Path, path.Method, op)
			err := refResolver.Resolve(op)
			if err != nil {
				return nil, err
			}
		}
	}
	return output, nil
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
			op.Operation.Extensions = make(map[string]interface{}, 6)
		}
		op.Operation.Extensions[vervet.ExtSnykApiResource] = op.ResourceName
		op.Operation.Extensions[vervet.ExtSnykApiVersion] = op.Version.String()
		op.Operation.Extensions[vervet.ExtSnykApiReleases] = releases
		op.Operation.Extensions[vervet.ExtSnykApiLifecycle] = op.Version.LifecycleAt(time.Time{}).String()
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
