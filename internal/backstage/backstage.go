// Package backstage supports vervet's integration with Backstage to
// automatically populate API definitions in the catalog info from compiled
// versions.
package backstage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"

	"github.com/snyk/vervet/v8"
)

const (
	backstageVersion   = "backstage.io/v1alpha1"
	snykApiVersionDate = "api.snyk.io/version-date"
	snykApiStability   = "api.snyk.io/version-stability"
	snykApiLifecycle   = "api.snyk.io/version-lifecycle"
	snykApiGeneratedBy = "api.snyk.io/generated-by"
)

// Component represents a Backstage Component entity document.
type Component struct {
	APIVersion string        `json:"apiVersion" yaml:"apiVersion"`
	Kind       string        `json:"kind" yaml:"kind"`
	Metadata   Metadata      `json:"metadata" yaml:"metadata"`
	Spec       ComponentSpec `json:"spec" yaml:"spec"`
}

// ComponentSpec represents a Backstage Component entity spec.
type ComponentSpec struct {
	Type         string   `json:"type" yaml:"type"`
	Owner        string   `json:"owner" yaml:"owner"`
	ProvidesAPIs []string `json:"providesApis" yaml:"providesApis"`
}

// API represents a Backstage API entity document.
type API struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
	Spec       APISpec  `json:"spec" yaml:"spec"`
}

// Metadata represents Backstage entity metadata.
type Metadata struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Title       string            `json:"title,omitempty" yaml:"title,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// APISpec represents a Backstage API entity spec.
type APISpec struct {
	Type       string        `json:"type" yaml:"type"`
	Lifecycle  string        `json:"lifecycle" yaml:"lifecycle"`
	Owner      string        `json:"owner" yaml:"owner"`
	System     string        `json:"system,omitempty" yaml:"system,omitempty"`
	Definition DefinitionRef `json:"definition" yaml:"definition"`
}

// DefinitionRef represents a reference to a local file in the project.
type DefinitionRef struct {
	Text string `json:"$text" yaml:"$text"`
}

// CatalogInfo models the Backstage catalog-info.yaml file at the top-level of
// a project.
type CatalogInfo struct {
	service          *yaml.Node
	serviceComponent Component
	components       []*yaml.Node
	VervetAPIs       []*API
}

// Save writes the catalog info to a writer.
func (c *CatalogInfo) Save(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	docs := []*yaml.Node{}
	if c.service != nil {
		docs = append(docs, c.service)
	}
	docs = append(docs, c.components...)
	sort.Sort(vervetAPIs(c.VervetAPIs))
	for _, vervetAPI := range c.VervetAPIs {
		var doc yaml.Node
		if err := doc.Encode(vervetAPI); err != nil {
			return err
		}
		doc.HeadComment = "Generated by vervet, DO NOT EDIT"
		docs = append(docs, &doc)
	}
	for _, doc := range docs {
		if err := enc.Encode(doc); err != nil {
			return err
		}
	}
	return nil
}

type vervetAPIs []*API

// Len implements sort.Interface.
func (v vervetAPIs) Len() int { return len(v) }

// Less implements sort.Interface.
func (v vervetAPIs) Less(i, j int) bool { return v[i].Metadata.Name < v[j].Metadata.Name }

// Swap implements sort.Interface.
func (v vervetAPIs) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// LoadCatalogInfo loads a catalog info from a reader.
func LoadCatalogInfo(r io.Reader) (*CatalogInfo, error) {
	dec := yaml.NewDecoder(r)
	var nodes []*yaml.Node
	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}
	catalog := &CatalogInfo{}
	vervetAPINames := map[string]struct{}{}
	for _, node := range nodes {
		if ok, err := isServiceComponent(node); err != nil {
			return nil, err
		} else if ok {
			catalog.service = node
			if err := node.Decode(&catalog.serviceComponent); err != nil {
				return nil, err
			}
			continue
		}
		if ok, err := isVervetGenerated(node); err != nil {
			return nil, err
		} else {
			if !ok {
				catalog.components = append(catalog.components, node)
			} else {
				// Remove prior vervet API names from the service component so we can rebuild them
				var api API
				if err := node.Decode(&api); err != nil {
					return nil, err
				}
				if api.Kind == "API" {
					vervetAPINames[api.Metadata.Name] = struct{}{}
				}
			}
		}
	}
	if catalog.service != nil {
		var apiNames []string
		for _, apiName := range catalog.serviceComponent.Spec.ProvidesAPIs {
			// Preserve manually added entries, things that are NOT vervet APIs
			if _, ok := vervetAPINames[apiName]; !ok {
				apiNames = append(apiNames, apiName)
			}
		}
		catalog.serviceComponent.Spec.ProvidesAPIs = apiNames
	}
	return catalog, nil
}

// LoadVervetAPIs loads all the compiled versioned OpenAPI specs and adds them
// to the catalog as API components.
func (c *CatalogInfo) LoadVervetAPIs(root, versions string, pivotDate time.Time, apiName string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	versions, err = filepath.Abs(versions)
	if err != nil {
		return err
	}
	specFiles, err := fs.Glob(os.DirFS(versions), "*/spec.json")
	if err != nil {
		return err
	}

	// Determine API names, combining existing + generated API entities.
	apiUniqueNames := map[string]struct{}{}
	for _, name := range c.serviceComponent.Spec.ProvidesAPIs {
		apiUniqueNames[name] = struct{}{}
	}
	for _, specFile := range specFiles {
		doc, err := vervet.NewDocumentFile(filepath.Join(versions, specFile))
		if err != nil {
			return err
		}
		api, err := c.vervetAPI(doc, root, pivotDate, apiName)
		if err != nil {
			return err
		}
		if _, ok := apiUniqueNames[api.Metadata.Name]; ok {
			return fmt.Errorf(`
there are multiple apis named %s, only one will be available on Backstage.
To resolve this error change the Name attribute in one of the spec files.
Note names may be truncated to fit the Backstage 63 character limit`,
				api.Metadata.Name,
			)
		}
		c.VervetAPIs = append(c.VervetAPIs, api)
		apiUniqueNames[api.Metadata.Name] = struct{}{}
	}
	apiNames := []string{}
	for name := range apiUniqueNames {
		apiNames = append(apiNames, name)
	}
	sort.Strings(apiNames)

	// Update the existing component providesApis with combined list of API
	// names.
	specPath, err := yamlpath.NewPath("$..spec")
	if err != nil {
		return err
	}
	specNodes, err := specPath.Find(c.service)
	if err != nil {
		return err
	}
	if len(specNodes) == 0 {
		return errors.New("missing spec in Backstage service component")
	}
	providesApisPath, err := yamlpath.NewPath("$.providesApis")
	if err != nil {
		return err
	}
	providesApisNodes, err := providesApisPath.Find(specNodes[0])
	if err != nil {
		return err
	}
	if len(providesApisNodes) == 0 {
		providesApisNodes = []*yaml.Node{{Kind: yaml.SequenceNode}}
		specNodes[0].Content = append(specNodes[0].Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "providesApis"},
			providesApisNodes[0],
		)
	}
	err = providesApisNodes[0].Encode(apiNames)
	if err != nil {
		return err
	}
	c.serviceComponent.Spec.ProvidesAPIs = apiNames
	return nil
}

// vervetAPI adds an OpenAPI spec document to the catalog.
func (c *CatalogInfo) vervetAPI(doc *vervet.Document, root string, pivotDate time.Time, apiName string) (*API, error) {
	version, err := doc.Version()
	if err != nil {
		return nil, err
	}
	ref, err := filepath.Rel(root, doc.Location().String())
	if err != nil {
		return nil, err
	}

	apiNameBS := toBackstageName(apiName)
	docTitleBS := toBackstageName(doc.Info.Title)
	dateStr := version.DateString()
	nameSuffix := fmt.Sprintf("_%s_%s", apiNameBS, dateStr)

	title := doc.Info.Title + " " + dateStr
	labels := map[string]string{
		snykApiVersionDate: dateStr,
	}
	tags := []string{version.Date.Format("2006-01")}
	specLifecycle := "production"

	// Specs generated after the pivot date have per operation stability, so
	// there is no global stability for the whole document. To preserve
	// backwards compatibility we still output metadata for the older specs.
	var name string
	if version.Date.Before(pivotDate) {
		lifecycle := version.LifecycleAt(time.Time{})
		var backstageLifecycle string
		if lifecycle == vervet.LifecycleReleased {
			backstageLifecycle = version.Stability.String()
		} else {
			backstageLifecycle = lifecycle.String()
		}
		specLifecycle = backstageLifecycle

		stabilityStr := version.Stability.String()
		stabilitySuffix := "_" + stabilityStr

		// Backstage names can only be a maximum of 63 characters
		availableTitleLen := 63 - len(nameSuffix) - len(stabilitySuffix)
		if availableTitleLen < 0 {
			availableTitleLen = 0
		}
		name = fmt.Sprintf("%."+strconv.Itoa(availableTitleLen)+"s%s%s", docTitleBS, nameSuffix, stabilitySuffix)
		title = title + " " + stabilityStr

		labels[snykApiStability] = stabilityStr
		labels[snykApiLifecycle] = lifecycle.String()
		tags = append(tags, stabilityStr, lifecycle.String())
	} else {
		// Backstage names can only be a maximum of 63 characters
		availableTitleLen := 63 - len(nameSuffix)
		name = fmt.Sprintf("%."+strconv.Itoa(availableTitleLen)+"s%s", docTitleBS, nameSuffix)
	}

	return &API{
		APIVersion: backstageVersion,
		Kind:       "API",
		Metadata: Metadata{
			Name:        name,
			Title:       title,
			Description: doc.Info.Description,
			Labels:      labels,
			Tags:        tags,
			Annotations: map[string]string{
				snykApiGeneratedBy: "vervet",
			},
		},
		Spec: APISpec{
			Type:      "openapi",
			Owner:     c.serviceComponent.Spec.Owner,
			Lifecycle: specLifecycle,
			Definition: DefinitionRef{
				Text: ref,
			},
		},
	}, nil
}

func toBackstageName(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r
		}
		if r >= 'a' && r <= 'z' {
			return r
		}
		if r == ' ' || r == '_' || r == '-' {
			return '-'
		}
		return -1
	}, strings.TrimSpace(s))
}

// isServiceComponent returns whether the YAML node is a Backstage component
// document for a service.
func isServiceComponent(node *yaml.Node) (bool, error) {
	var doc Component
	if err := node.Decode(&doc); err != nil {
		return false, err
	}
	return doc.Kind == "Component", nil
}

// isVervetGenerated returns whether the YAML node is a Backstage entity
// document that was generated by Vervet.
func isVervetGenerated(node *yaml.Node) (bool, error) {
	var comp Component
	if err := node.Decode(&comp); err != nil {
		return false, err
	}
	return comp.Metadata.Annotations[snykApiGeneratedBy] == "vervet", nil
}
