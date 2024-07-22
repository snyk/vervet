package simplebuild

import (
	"context"
	"fmt"
	"path/filepath"

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
		err = docs.Write()
		if err != nil {
			return err
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
	Version vervet.Version
	Doc     *openapi3.T
}
type DocSet []VersionedDoc

func (ops Operations) Build() (DocSet, error) {
	versionIndex := ops.Versions()
	versions := versionIndex.Versions()
	output := make(DocSet, len(versions))
	for idx, version := range versions {
		output[idx] = VersionedDoc{
			Doc:     &openapi3.T{},
			Version: version,
		}
		for path, spec := range ops {
			op := spec.GetLatest(version)
			if op == nil {
				continue
			}
			output[idx].Doc.AddOperation(path.Path, path.Method, op)
		}
	}
	return output, nil
}

func (ops Operations) Versions() vervet.VersionIndex {
	versionSet := map[vervet.Version]struct{}{}
	for _, opSet := range ops {
		for _, op := range opSet {
			versionSet[op.Version] = struct{}{}
		}
	}
	uniqueVersions := make(vervet.VersionSlice, len(versionSet))
	idx := 0
	for version := range versionSet {
		uniqueVersions[idx] = version
		idx++
	}
	return vervet.NewVersionIndex(uniqueVersions)
}

func (docs DocSet) Write() error {
	for _, doc := range docs {
		fmt.Println(doc.Version)
		out, err := doc.Doc.MarshalJSON()
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	}
	return nil
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

func (vs VersionSet) GetLatest(before vervet.Version) *openapi3.Operation {
	var latest *VersionedOp
	for _, versionedOp := range vs {
		isBefore := versionedOp.Version.Compare(before) <= 0
		isLowerStability := versionedOp.Version.Stability.Compare(before.Stability) < 0
		if isBefore && !isLowerStability {
			if latest == nil || versionedOp.Version.Compare(latest.Version) > 0 {
				latest = &versionedOp
			}
		}
	}
	if latest == nil {
		return nil
	}
	return latest.Operation
}
