package simplebuild

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/config"
	"github.com/snyk/vervet/v7/internal/files"
)

func Build(ctx context.Context, project *config.Project) error {
	for _, apiConfig := range project.APIs {
		fmt.Printf("Processing API: %s\n", apiConfig.Name)

		operations, err := LoadPaths(ctx, apiConfig)
		if err != nil {
			return err
		}
		for _, op := range operations {
			op.Annotate()
		}
		docs, err := operations.Build()
		if err != nil {
			return err
		}

		err = docs.ApplyOverlays(ctx, apiConfig.Overlays)
		if err != nil {
			return err
		}

		sortDocsByVersionDate(docs)

		err = CheckBreakingChanges(docs)
		if err != nil {
			return err
		}

		if apiConfig.Output != nil {
			err = docs.WriteOutputs(*apiConfig.Output)
			if err != nil {
				return err
			}
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

func (ops Operations) Build() (DocSet, error) {
	versionDates := ops.VersionDates()
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

			doc.InternalizeRefs(ctx, nil)
			err = doc.ResolveRefs()
			if err != nil {
				return nil, fmt.Errorf("failed to localize refs: %w", err)
			}

			for pathName, pathDef := range doc.T.Paths {
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

func CheckBreakingChanges(docs DocSet) error {
	for i := 1; i < len(docs); i++ {
		prevDoc := docs[i-1]
		currDoc := docs[i]

		// Create temporary file paths for previous and current specs
		prevSpecPath := filepath.Join(os.TempDir(), fmt.Sprintf("spec-%d.json", i-1))
		currSpecPath := filepath.Join(os.TempDir(), fmt.Sprintf("spec-%d.json", i))

		// Write specs to temporary files
		if err := writeTempSpecFile(prevDoc.Doc, prevSpecPath); err != nil {
			return err
		}
		if err := writeTempSpecFile(currDoc.Doc, currSpecPath); err != nil {
			return err
		}

		cmd := exec.Command("oasdiff", "breaking", "--fail-on", "ERR", prevSpecPath, currSpecPath)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return fmt.Errorf("no breaking change detected between versions %s and %s: \n %s",
				prevDoc.VersionDate.Format(time.DateOnly), currDoc.VersionDate.Format(time.DateOnly), string(output))
		}

		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			continue // breaking change detected, continue to the next pair
		}
		if os.IsNotExist(err) {
			return fmt.Errorf("oasdiff executable not found. Please ensure it is installed and available in your PATH")
		}
		return fmt.Errorf("failed to run oasdiff: %w", err)
	}
	return nil
}

func writeTempSpecFile(doc *openapi3.T, path string) error {
	jsonBuf, err := doc.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal spec to JSON: %w", err)
	}
	return os.WriteFile(path, jsonBuf, 0644)
}
