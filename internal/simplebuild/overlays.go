package simplebuild

import (
	"context"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/config"
)

func (docs DocSet) ApplyOverlays(ctx context.Context, cfgs []*config.Overlay) error {
	overlays, err := loadOverlays(ctx, cfgs)
	if err != nil {
		return fmt.Errorf("load overlays: %w", err)
	}
	for _, doc := range docs {
		for _, overlay := range overlays {
			// NB: Will overwrite any existing definitions without warning.
			err := vervet.Merge(doc.Doc, overlay, true)
			if err != nil {
				return fmt.Errorf("apply overlay: %w", err)
			}
		}
	}

	return nil
}

func loadOverlays(ctx context.Context, cfgs []*config.Overlay) ([]*openapi3.T, error) {
	overlays := make([]*openapi3.T, len(cfgs))
	for idx, overlayCfg := range cfgs {
		if overlayCfg.Include != "" {
			doc, err := vervet.NewDocumentFile(overlayCfg.Include)
			if err != nil {
				return nil, fmt.Errorf("load include overlay: %w", err)
			}
			err = vervet.Localize(ctx, doc)
			if err != nil {
				return nil, fmt.Errorf("localise overlay: %w", err)
			}
			overlays[idx] = doc.T
		} else if overlayCfg.Inline != "" {
			docString := os.ExpandEnv(overlayCfg.Inline)
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte(docString))
			if err != nil {
				return nil, fmt.Errorf("load inline overlay: %w", err)
			}
			overlays[idx] = doc
		}
	}
	return overlays, nil
}
