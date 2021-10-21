# vervet

Vervet is an HTTP API version lifecycle management tool, allowing APIs to be designed, developed, versioned and released from [resources](https://github.com/snyk/sweater-comb/blob/main/docs/intro.md#resources) independently and concurrently.

In a large organization, there might be many teams involved in delivering a large API -- such as at [Snyk](https://snyk.io) where Vervet was developed.

Within a single small team, there is still often a need to simultaneously try new things in parts of an API while maintaining stability.

While Vervet was developed in the context of a RESTful API, Vervet can be used with any HTTP API expressed in OpenAPI 3 -- even if it does not adhere to strict REST principles.

### [API Versioning](https://github.com/snyk/sweater-comb/blob/main/docs/version.md)

To summarize the API versioning supported by Vervet:

#### What is versioned?
Resource versions are defined in OpenAPI 3, as if the resource were a standalone service.

#### How are resource version specs organized?
Resources are organized in a standard directory structure by release date, using OpenAPI extensions to define lifecycle concepts like stability.

#### How does versioning work?
* Resources are versioned independently by date and stability, with a well-defined deprecation and sunsetting policy.
* Additive, non-breaking changes can be made to released versions. Breaking changes trigger a new version.
* New versions deprecate and sunset prior versions, on a timeline determined by the stability level.

[Read more about API versioning](https://github.com/snyk/sweater-comb/blob/main/docs/version.md).

## Features

A brief tour of Vervet's features.

### Compilation

Vervet compiles the OpenAPI spec of each resource version into a series of OpenAPI specifications that describe the entire application, at each distinct release in its underlying parts.

Given a directory structure of resource versions, each defined by an OpenAPI spec as if it were an independent service:

```
$ tree resources
resources
├── _examples
│   └── hello-world
│       ├── 2021-06-01
│       │   └── spec.yaml
│       ├── 2021-06-07
│       │   └── spec.yaml
│       └── 2021-06-13
│           └── spec.yaml
└── projects
    └── 2021-06-04
        └── spec.yaml
```

and a Vervet project configuration that instructs how to put them together:

```yml
$ cat .vervet.yaml
apis:
  my-api:
    resources:
      - path: 'resources'
    output:
      path: 'versions'
```

`vervet compile` aggregates these resources' individual OpenAPI specifications to describe the entire service API _at each distinct version date and stability level_ from its component parts.

```
$ tree versions
versions/
├── 2021-06-01
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-01~beta
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-01~experimental
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-04
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-04~beta
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-04~experimental
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-07
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-07~beta
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-07~experimental
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-13
│   ├── spec.json
│   └── spec.yaml
├── 2021-06-13~beta
│   ├── spec.json
│   └── spec.yaml
└── 2021-06-13~experimental
    ├── spec.json
    └── spec.yaml
```

### Linting

Vervet is not an OpenAPI linter. It coordinates and frontends OpenAPI linting, allowing different rules to be applied to different parts of an API, or different stages of the compilation process (source component specs, output compiled specs). It also allows exceptions to be made to certain resource versions, so that new rules do not break already-released parts of the API.

Vervet currently supports linting OpenAPI specifications with:
* [Spectral](https://stoplight.io/open-source/spectral/)
* [Sweater Comb](https://github.com/snyk/sweater-comb), as a self-contained Docker image which combines a linter and custom opinionated rulesets.

Direct Spectral linting may be soon deprecated in favor of container-based linting.

### Generation

Since Vervet models the composition and construction of an API, it is well positioned to coordinate code and artifact generation through templates.

Generators are defined in `.vervet.yaml`:

```yml
generators:
  version-readme:
    scope: version
    filename: "resources/{{ .Resource }}/{{ .Version }}/README"
    template: ".vervet/templates/README.tmpl"
  version-spec:
    scope: version
    filename: "resources/{{ .Resource }}/{{ .Version }}/spec.yaml"
    template: ".vervet/templates/spec.yaml.tmpl"
```

In this case, generators produce a boilerplate OpenAPI specification containing HTTP methods to create, list, get, update, and delete a resource, and a nice README when a new resource version is created. OpenAPI specifications can be tedious to write from scratch; generators help developers focus on adding the content that matters most.

Generators are defined using [Go templates](https://pkg.go.dev/text/template). Template syntax is also used to express filename interpolation per resource, per version.

```yml
apis:
  my-api:
    resources:
      - path: 'resources'
    generators:
      - version-readme
      - version-spec
    output:
      path: 'versions'
```

Generators are applied during lifecycle commands, such as creating a new resource version:

```
$ vervet version new my-api thing
$ tree resources
resources
└── thing
    └── 2021-10-21
        ├── README
        └── spec.yaml
```

Generators support multiple stages. For example, once a boilerplate spec.yaml is generated, it can be fed into subsequent generators that produce code, API gateway configuration, Grafana dashboards, and HTTP load tests.

A more advanced example, ExpressJS controllers generated from each operation in a resource version OpenAPI spec:

```yml
generators:
  version-spec:
    scope: version
    filename: "resources/{{ .Resource }}/{{ .Version }}/spec.yaml"
    template: ".vervet/templates/spec.yaml.tmpl"
  version-controller:
    scope: version
    # `files:` generates a collection of files -- which itself is expressed as a
    # YAML template.  Keys in this YAML are the paths of the files to generate,
    # whose values are the file contents.
    files: |-
      {{- $resource := .Resource -}}
      {{- $version := .Version -}}
      {{- range $path, $pathItem := .Data.Spec.paths -}}
      {{- range $method, $operation := $pathItem -}}
      {{- $operationId := $operation.operationId -}}
      {{/* Construct a context object using the 'map' function */}}
      {{- $ctx := map "Context" . "OperationId" $operationId }}
      resources/{{ $resource }}/{{ $version }}/{{ $operationId }}.ts: |-
        {{/*
             Evaluate the template by including it with the necessary context.
             The generator's template is included as "contents" from within the
             `files:` template.
           */}}
        {{ include "contents" $ctx | indent 2 }}
      {{ end }}
      {{- end -}}
    template: ".vervet/resource/version/controller.ts.tmpl"
    data:
      Spec:
        # generated above in version-spec, accessible from within the `files:`
        # template as `.Data.Spec`.
        include: "resources/{{ .Resource }}/{{ .Version }}/spec.yaml"
apis:
  my-api:
    resources:
      - path: 'resources'
        generators:
          # order is important
          - version-spec
          - version-controller
    output:
      path: 'versions'
```

In this case, a template is being applied per `operationId` in the `spec.yaml` generated in the prior step. `version-controller` produces a collection of files, a controller module per resource, per version, per operation. This is possible because generators are applied in the order they are declared on each set of resources.

### Scaffolding

Just as generators automate the generation of artifacts as part of the versioning lifecycle, scaffolds are used to bootstrap a new greenfield Vervet API project with useful defaults:

* Vervet project configuration (`.vervet.yaml`)
* Directory structure and layout for API specifications
* Generator templates
* Linter rulesets

Scaffolds are great in a microservice/SOA self-service ecosystem, where new services may be created often, and need a set of sensible defaults to quickly get started.

```
$ mkdir my-new-service
$ cd my-new-service
$ vervet scaffold init ../vervet-api-scaffold/
$ tree -a
.
├── .vervet
│   ├── components
│   │   ├── common.yaml
│   │   ├── errors.yaml
│   │   ├── headers
│   │   │   └── headers.yaml
│   │   ├── parameters
│   │   │   ├── pagination.yaml
│   │   │   └── version.yaml
│   │   ├── responses
│   │   │   ├── 204.yaml
│   │   │   ├── 400.yaml
│   │   │   ├── 401.yaml
│   │   │   ├── 403.yaml
│   │   │   ├── 404.yaml
│   │   │   ├── 409.yaml
│   │   │   ├── 429.yaml
│   │   │   └── 500.yaml
│   │   ├── tag.yaml
│   │   ├── types.yaml
│   │   └── version.yaml
│   ├── openapi
│   │   └── spec.yaml
│   └── templates
│       ├── README.tmpl
│       └── spec.yaml.tmpl
├── .vervet.yaml
└── api
    ├── resources
    └── versions
```

This scaffold sets up a new project with standard OpenAPI components that are referenced by resource OpenAPI boilerplate templates. New resources are generated already conforming to our [JSON API](https://github.com/snyk/sweater-comb/blob/main/docs/jsonapi.md) standards and paginated list operations.

## Installation

### NPM

    npm install -g @snyk/vervet

NPM packaging adapted from https://github.com/manifoldco/torus-cli.

### Source

Go >= 1.16 required.

    go build ./cmd/vervet

or

    make build

## Development

Vervet uses a reference set of OpenAPI documents in `testdata/resources` in
tests. CLI tests compare runtime compiled output with pre-compiled, expected
output in `testdata/output` to detect regressions.

When introducing changes that intentionally change the content of compiled
output:

* Run `go generate ./testdata` to update the contents of `testdata/output`
* Verify that the compiled output is correct
* Commit the changes to `testdata/output` in your proposed branch
