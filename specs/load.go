package specs

import (
	"context"
	"fmt"
	"iter"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/internal/files"
)

// GetInputSpecs returns a list of all of the input specs for a given project.
// It will resolve all references in the spec so the output can be consumed by
// tooling unaware applications.
func GetInputSpecs(ctx context.Context, api *config.API) ([]*vervet.Document, error) {
	docs := []*vervet.Document{}
	for doc, err := range GetInputSpecsItr(ctx, api) {
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// GetInputSpecsItr is a iterator version of GetInputSpecs. It is preferred for
// lazy operations.
func GetInputSpecsItr(ctx context.Context, api *config.API) iter.Seq2[*vervet.Document, error] {
	return func(yield func(*vervet.Document, error) bool) {
		for _, resource := range api.Resources {
			paths, err := files.LocalFSSource{}.Match(resource)
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}
			for _, path := range paths {
				doc, err := vervet.NewDocumentFile(path)
				if err != nil {
					if !yield(nil, fmt.Errorf("failed to load spec: %w", err)) {
						return
					}
				}
				doc.InternalizeRefs(ctx, vervet.ResolveRefsWithoutSourceName)
				err = doc.ResolveRefs()
				if !yield(doc, err) {
					return
				}
			}
		}
	}
}
