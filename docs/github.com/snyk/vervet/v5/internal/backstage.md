# backstage

```go
import "github.com/snyk/vervet/v5/internal/backstage"
```

Package backstage supports vervet's integration with Backstage to automatically populate API definitions in the catalog info from compiled versions\.

## Index

- [type API](<#type-api>)
- [type APISpec](<#type-apispec>)
- [type CatalogInfo](<#type-cataloginfo>)
  - [func LoadCatalogInfo(r io.Reader) (*CatalogInfo, error)](<#func-loadcataloginfo>)
  - [func (c *CatalogInfo) LoadVervetAPIs(root, versions string) error](<#func-cataloginfo-loadvervetapis>)
  - [func (c *CatalogInfo) Save(w io.Writer) error](<#func-cataloginfo-save>)
- [type Component](<#type-component>)
- [type ComponentSpec](<#type-componentspec>)
- [type DefinitionRef](<#type-definitionref>)
- [type Metadata](<#type-metadata>)


## type [API](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L46-L51>)

API represents a Backstage API entity document\.

```go
type API struct {
    APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
    Kind       string   `json:"kind" yaml:"kind"`
    Metadata   Metadata `json:"metadata" yaml:"metadata"`
    Spec       APISpec  `json:"spec" yaml:"spec"`
}
```

## type [APISpec](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L65-L71>)

APISpec represents a Backstage API entity spec\.

```go
type APISpec struct {
    Type       string        `json:"type" yaml:"type"`
    Lifecycle  string        `json:"lifecycle" yaml:"lifecycle"`
    Owner      string        `json:"owner" yaml:"owner"`
    System     string        `json:"system,omitempty" yaml:"system,omitempty"`
    Definition DefinitionRef `json:"definition" yaml:"definition"`
}
```

## type [CatalogInfo](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L80-L85>)

CatalogInfo models the Backstage catalog\-info\.yaml file at the top\-level of a project\.

```go
type CatalogInfo struct {
    VervetAPIs []*API
    // contains filtered or unexported fields
}
```

### func [LoadCatalogInfo](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L125>)

```go
func LoadCatalogInfo(r io.Reader) (*CatalogInfo, error)
```

LoadCatalogInfo loads a catalog info from a reader\.

### func \(\*CatalogInfo\) [LoadVervetAPIs](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L181>)

```go
func (c *CatalogInfo) LoadVervetAPIs(root, versions string) error
```

LoadVervetAPIs loads all the compiled versioned OpenAPI specs and adds them to the catalog as API components\.

### func \(\*CatalogInfo\) [Save](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L88>)

```go
func (c *CatalogInfo) Save(w io.Writer) error
```

Save writes the catalog info to a writer\.

## type [Component](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L31-L36>)

Component represents a Backstage Component entity document\.

```go
type Component struct {
    APIVersion string        `json:"apiVersion" yaml:"apiVersion"`
    Kind       string        `json:"kind" yaml:"kind"`
    Metadata   Metadata      `json:"metadata" yaml:"metadata"`
    Spec       ComponentSpec `json:"spec" yaml:"spec"`
}
```

## type [ComponentSpec](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L39-L43>)

ComponentSpec represents a Backstage Component entity spec\.

```go
type ComponentSpec struct {
    Type         string   `json:"type" yaml:"type"`
    Owner        string   `json:"owner" yaml:"owner"`
    ProvidesAPIs []string `json:"providesApis" yaml:"providesApis"`
}
```

## type [DefinitionRef](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L74-L76>)

DefinitionRef represents a reference to a local file in the project\.

```go
type DefinitionRef struct {
    Text string `json:"$text" yaml:"$text"`
}
```

## type [Metadata](<https://github.com/snyk/vervet/blob/main/internal/backstage/backstage.go#L54-L62>)

Metadata represents Backstage entity metadata\.

```go
type Metadata struct {
    Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
    Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
    Title       string            `json:"title,omitempty" yaml:"title,omitempty"`
    Description string            `json:"description,omitempty" yaml:"description,omitempty"`
    Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
    Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
    Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
}
```

