package simplebuild

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/config"
	"github.com/snyk/vervet/v7/internal/files"
)

func Build(ctx context.Context, project *config.Project) error {
	for _, apiConfig := range project.APIs {
		operations, err := LoadPaths(ctx, apiConfig)
		if err != nil {
			return err
		}
		docs, err := operations.Build()
		if err != nil {
			return err
		}

		docs.ApplyOverlays(ctx, apiConfig.Overlays)

		if apiConfig.Output != nil {
			err = docs.WriteOutputs(*apiConfig.Output)
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
	Version   vervet.Version
	Operation *openapi3.Operation
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
						Version:   version,
						Operation: opDef,
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
