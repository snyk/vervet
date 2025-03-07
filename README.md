# Vervet

Vervet is an HTTP API version lifecycle management tool, allowing APIs to be designed, developed, versioned and released from [resources](https://github.com/snyk/sweater-comb/blob/main/docs/principles/api_program.md#resources) independently and concurrently.

In a large organization, there might be many teams involved in delivering a large API -- such as at [Snyk](https://snyk.io) where Vervet was developed.

Within a single small team, there is still often a need to simultaneously try new things in parts of an API while maintaining stability.

While Vervet was developed in the context of a RESTful API, Vervet can be used with any HTTP API expressed in OpenAPI 3 -- even if it does not adhere to strict REST principles.

### [API Versioning](https://github.com/snyk/sweater-comb/blob/main/docs/principles/version.md)

To summarize the API versioning supported by Vervet:

#### What is versioned?

Resource versions are defined in OpenAPI 3, as if each resource were a standalone service.

#### How are resource version specs organized?

Resources are organized in a standard directory structure by release date, using OpenAPI extensions to define lifecycle concepts like stability.

#### How does versioning work?

- Resources are versioned independently by date and stability, with a well-defined deprecation and sunsetting policy.
- Additive, non-breaking changes can be made to released versions. Breaking changes trigger a new version.
- New versions deprecate and sunset prior versions, on a timeline determined by the stability level.

[Read more about API versioning](https://github.com/snyk/sweater-comb/blob/main/docs/principles/version.md).

## Features

A brief tour of Vervet's features.

### Building a service OpenAPI from resources

Vervet collects the OpenAPI specification of each resource version and merges them into a series of OpenAPI specifications that describe the entire application, at each distinct release version in its underlying parts.

Given a directory structure of resource versions, each defined by an OpenAPI specification as if it were an independent service:

    tree resources

```
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

    cat .vervet.yaml

```yaml
apis:
  my-api:
    resources:
      - path: "resources"
    output:
      path: "versions"
```

`vervet build` aggregates these resources' individual OpenAPI specifications to describe the entire service API _at each distinct version date and stability level_ from its component parts.

    tree versions

```
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

### Simplified Versioning (from 2024-10-15)

From 2024-10-15, Vervet introduced a new "simplified versioning" scheme.

The main differences introduced by simplified versioning are:

- **Stability dropped from the spec level to the individual path level**: Each API path can now be labeled as either `beta` or `GA` (general availability). The `experimental` stability level has been removed. This means that paths within the same version of the spec can have different stability statuses, allowing for greater flexibility and precision.

- **No changes required from developers defining APIs**: Developers still define their APIs in the same way as before, and Vervet handles the collation and versioning of specs in the new format automatically. The way specs are defined, updated, and maintained remains familiar, ensuring a smooth transition.

- **End-user API requests are simplified**: API consumers no longer need to specify the stability when calling a particular version. They can simply provide the date of the version they want, and Vervet will resolve whether each path is `GA` or `beta`. If a path has been promoted to `GA`, subsequent versions cannot publish it as `beta` again.

This change is aimed at simplifying how users interact with versioned APIs while maintaining clear definitions and stabilities at the path level.

#### Examples for Simplified Versioning

To illustrate how simplified versioning works in practice, let’s consider some examples:

##### Example 1: Requesting a Version without Specifying Stability

Consider an API with paths `/pets` and `/owners`. Suppose the `/pets` path is `GA` and the `/owners` path is `beta` as of version date `2024-10-20`.

- **Request:**

  ```
  GET /api/2024-10-20/pets
  ```

  - **Response:** This request will be handled by the GA version of `/pets`.

- **Request:**
  ```
  GET /api/2024-10-20/owners
  ```
  - **Response:** This request will be handled by the beta version of `/owners`.

The user does not need to specify the stability explicitly. Vervet determines the appropriate path stability (`GA` or `beta`) automatically.

##### Example 2: Path Promotion to GA

Let’s say that as of `2024-11-10`, the `/owners` path is promoted to `GA`.

- **Request:**
  ```
  GET /api/2024-11-15/owners
  ```
  - **Response:** The request will be handled by the GA version of `/owners`.

Once a path is promoted to GA, any subsequent version cannot publish it as `beta` again. Therefore, users can be confident that they are always accessing a stable version of a path if it is marked as GA.

##### Example 3: Multiple Paths with Mixed Stability

Consider an API version dated `2024-12-01` with the following paths:

- `/pets`: GA
- `/owners`: GA
- `/appointments`: beta

- **Request:**

  ```
  GET /api/2024-12-01/appointments
  ```

  - **Response:** The request will be handled by the beta version of `/appointments`.

- **Request:**
  ```
  GET /api/2024-12-01/pets
  ```
  - **Response:** The request will be handled by the GA version of `/pets`.

This approach allows different parts of an API to evolve at different paces, providing flexibility for developers and a clearer experience for end-users.

## Code generation

Since Vervet models the composition, construction and versioning of an API, it is well positioned to coordinate code and artifact generation through the use of templates.

Generators may be defined in a YAML file, such as `generators.yaml`:

```yaml
generators:
  version-readme:
    scope: version
    filename: "{{ .Path }}/README"
    template: "{{ .Here }}/templates/README.tmpl" # Located relative to the location of generators.yaml
```

The context of `README.tmpl` has full access to the resource version metadata and OpenAPI document object model.

```yaml
Generated by vervet. DO NOT EDIT!

# My API

Files in this directory were generated by `@snyk/vervet`
for resource `{{ .ResourceVersion.Name }}` at version `{{ .ResourceVersion.Version.String }}`.
```

In a project with a `.vervet.yaml` configuration, execute the generators with

    vervet generate -g generators.yaml

The simple generator above produces a README in each resource version directory.

    tree resources

```
resources
└── thing
    └── 2021-10-21
        ├── README
        └── spec.yaml
```

Generators are defined using [Go templates](https://pkg.go.dev/text/template).

Template syntax may also be used to express a directory structure of many files. A more advanced example, an Express controller generated from each operation in a resource version OpenAPI spec:

```yaml
generators:
  version-controller:
    scope: version
    # `files:` generates a collection of files -- which itself is expressed as a
    # YAML template.  Keys in this YAML are the paths of the files to generate,
    # whose values are the file contents.
    files: |-
      {{- $path := .Path -}}
      {{- $resource := .ResourceVersion -}}
      {{- $version := .ResourceVersion.Version -}}
      {{- range $path, $pathItem := .ResourceVersion.Document.Paths -}}
      {{- range $method, $operation := $pathItem -}}
      {{- $operationId := $operation.operationId -}}
      {{/* Construct a context object using the 'map' function */}}
      {{- $ctx := map "Context" . "OperationId" $operationId }}
      {{ $path }}/{{ $operationId }}.ts: |-
        {{/*
             Evaluate the template by including it with the necessary context.
             The generator's template (controller.ts.tmpl) is included as
             "contents" from within the `files:` template.
           */}}
        {{ include "contents" $ctx | indent 2 }}
      {{ end }}
      {{- end -}}
    template: "{{ .Here }}/templates/controller.ts.tmpl"
```

In this case, a template is being applied per `operationId` in the `spec.yaml` generated in the prior step. `version-controller` produces a collection of files, a controller module per resource, per version, per operation.

Finally, a note on scoping. Generators can be scoped to either a `version` or a `resource`.

`scope: version` generator templates execute with [VersionScope](https://pkg.go.dev/github.com/snyk/vervet/v6/internal/generator#VersionScope). This maps 1:1 with a single resource version OpenAPI specification.

`scope: resource` generator templates execute with [ResourceScope](https://pkg.go.dev/github.com/snyk/vervet/v6/internal/generator#ResourceScope). This is a collection of resource versions, useful for building resource routers.

## Installation

### NPM

Within a project:

    npm install @snyk/vervet

Or installed globally:

    npm install -g @snyk/vervet

NPM packaging adapted from https://github.com/manifoldco/torus-cli.

### Go

Go >= 1.16 required.

    go install github.com/snyk/vervet/v6/cmd/vervet@latest

Building from source locally:

    go build ./cmd/vervet

or

    make build

## Development

Vervet uses a reference set of OpenAPI documents in `testdata/resources` in
tests. CLI tests compare runtime compiled output with pre-compiled, expected
output in `testdata/output` to detect regressions.

When introducing changes that intentionally change the content of compiled
output:

- Run `go generate ./testdata` to update the contents of `testdata/output`
- Verify that the compiled output is correct
- Commit the changes to `testdata/output` in your proposed branch

## Releasing a new version

A new version of `vervet` will automatically be generated for Github and `npm` when new features
are introduced, i.e. when commits are merged that are marked with `feat:`.

## Deprecating a version

After removing the endpoint version code and specs, you may see this issue:

```
ENOENT: no such file or directory, open '.../spec.yaml'
```

To solve this:

1. Temporarily ignore the endpoint version code in `.vervet.yaml`
2. Remove the endpoint versions from `catalog-info.yaml`
3. Remove the old OpenAPI specs.

[Example PR](https://github.com/snyk/registry/pull/33489/files)

# Vervet Underground

# What Vervet Underground does and why

In order to understand _why_ Vervet Underground exists and the problem it solves, you should first become familiar with the [API versioning scheme](https://github.com/snyk/sweater-comb/blob/main/docs/principles/version.md) that Vervet supports. The main idea is, an API may be authored in parts, each of those parts may be versioned, and all the distinct versions are assembled to produce a cohesive timeline of versions for the entire service.

Just as Vervet compiles a timeline of OpenAPI versions for a single service from independently versioned parts, Vervet Underground (VU) compiles a timeline of OpenAPI spec versions for a SaaS from independently versioned microservices, each of which contributes parts of the SaaS API.

# Service API aggregation by example

## The pet store gets microservices

To illustrate how this works in practice, let's deconstruct a pet store into two services:

- `petfood.default.svc.cluster.local`, which knows about pet food.
- `animals.default.svc.cluster.local`, which knows about animals.

For sake of example, let's assume the following versions are published by each service:

petfood has:

- `2021-07-04~experimental`
- `2021-08-09~beta`
- `2021-08-09~experimental` (beta released, with some parts still experimental, so both are published)
- `2021-09-14` . (first GA)
- `2021-09-14~beta`
- `2021-09-14~experimental`

animals has:

- `2021-09-10~experimental`
- `2021-10-04~experimental`
- `2021-10-12~beta`
- `2021-10-12~experimental`
- `2021-11-05`
- `2021-11-05~beta`
- `2021-11-05~experimental`

And, the OpenAPI spec for each version is available at `/openapi`. `/openapi` provides a JSON array of OpenAPI versions supported by the service, and `/openapi/{version}` fetches the OpenAPI spec for that particular `{version}`. For example, `GET http://petfood.default.svc.cluster.local/openapi/2021-09-14~experimental`.

There is some nuance to this to be aware of. You'll notice that some dates have multiple versions with different stabilities. This can happen because on that date, there is more than one API version available at different stability levels.

There are also some assumptions. These services are cooperatively contributing to the public pet store SaaS API. They cannot conflict with each other -- no overwriting each other's OpenAPI components, or publishing conflicting paths.

## Pet store's public API

From these service versions, what versions of the pet store API are published to the public SaaS consumer? Well, the union of all of them! So we should be able to enumerate these versions from the public SaaS API:

```
GET https://api.petstore.example.com/openapi

200 OK
Content-Type: application/json
[
  `2021-07-04~experimental`,   // Contains only petfood so far...
  `2021-08-09~experimental`,
  `2021-08-09~beta`,
  `2021-09-10~beta`            // Petfood only (no beta version of animals yet...)
  `2021-09-10~experimental`,   // First animals (experimental version) + petfood
  `2021-09-14`,                // Petfood GA only, animals isn't GA yet
  `2021-09-14~beta`,
  `2021-09-14~experimental`,
  `2021-10-04`,
  `2021-10-04~beta`,
  `2021-10-04~experimental`,
  `2021-10-12`,
  `2021-10-12~beta`,
  `2021-10-12~experimental`,
  `2021-11-05`,                // First GA release of animals, also has petfood GA (from most recent 2021-09-14)
  `2021-11-05~beta`,
  `2021-11-05~experimental`,
]
```

## Past API releases can change

The examples so far have been kept simple by assuming released API versions do not change. In practice, non-breaking changes are allowed to be made to existing versions of the API at any time. [Non-breaking changes](https://github.com/snyk/sweater-comb/blob/main/docs/principles/version.md#breaking-changes) must be additive and optional. It is fine to add new HTTP methods, endpoints, request parameters or response fields, so long that the added parameters or fields are not _required_ -- which existing generated client code would have no way of knowing about.

With VU, we should be able to request a version of the public SaaS API _as it was on a given date_, regardless of the version release date.

### Tracking a non-breaking change by example

Let's assume that on 2021-11-08, the petfood service team adds a PATCH method to all their resources, to allow existing orders to be modified before they ship. It's a reasonable thing to do -- a new HTTP method doesn't break existing behavior! The team adds this method to every active (non-sunset) version retroactively -- why not? it was essentially the same backend code to implement it!

Vervet Underground not only compiles the initially published APIs from component services, it tracks changes in these APIs and updates its SaaS-level view of the API accordingly. So, VU scrapes the /openapi endpoints of its services periodically and detects the API changes on 2021-11-08, even though no new API was explicitly released that day, and now represents it as a new "discovered" version:

```
GET https://api.petstore.example.com/openapi

200 OK
Content-Type: application/json
[
  `2021-07-04~experimental`,
  ...
  `2021-11-05`,
  `2021-11-05~beta`,
  `2021-11-05~experimental`,
  `2021-11-08`,                // A wild new version appears!
  `2021-11-08~beta`,
  `2021-11-08~experimental`,
]
```

## How VU tracks non-breaking changes (even our versions have versions!)

VU scrapes the `/openapi` endpoints of each service and tracks the changes in each version found. This may be stored in a directory structure, such as:

```
services/petfood
├── 2021-09-14~experimental
│   ├── 2021-09-14_11_23_24.spec.yaml
│   └── 2021-11-08_13_14_15.spec.yaml
...
```

where the scraped OpenAPI is compared against the most recent last snapshot: if they differ, a new snapshot is taken.

When the new `2021-11-08` snapshot is detected, this triggers a rebuild of the top-level SaaS OpenAPI specs with that new version added. The arrow of time eventually flows only one way and storage is cheap, so it's assumed that the compiled OpenAPI specs will be statically compiled up-front as service API changes are detected.

This snapshot version should not be taken to represent a breaking-change release, which has different deprecation and sunsetting implications. It is only used to represent what the API looked like at a given point in time.

### What this means from a public API perspective

If a version date prior to `2021-11-08` resolves to `2021-09-14~experimental`, you should see the `2021-09-14~experimental` contributions to the API _as it would have appeared at that time_, in other words, you should see a view based on the `2021-07-04_11_23_24.spec.yaml` snapshot of `2021-09-14~experimental`.

If a version date after `2021-11-08` matches `2021-09-14~experimental`, let's say a request for `2021-12-10~experimental`, then you should see it as it would appear after the non-breaking change, `2021-11-08_13_14_15.spec.yaml`.

# Roadmap

## Minimum Viable

VU aggregates OpenAPI specs from multiple services and serves them up in a single place for:

- Docs
  - Docs will likely either render public `/openapi` directly client-side or periodically scrape & update
- Routing public v3 `/openapi` to VU's aggregated OpenAPI versions
- Add Akamai configuration to serve the blockstored specs initially for `/openapi`
  - Formal frontend presentation will be later

This unblocks docs for multi-service decomp.

### Details

- Could use block storage with a history of changes per service per version
- Static config that tells VU where to scrape upstream OpenAPI from services (registry and friends)
- Cron job to periodically scrape and update
  - Or we can try to make this push and set up a webhook...

##### Simplified Versioning Integration

From 2024-10-15, Vervet Underground also integrates the simplified versioning model. With simplified versioning:

- API versions compiled by VU no longer require consumers to specify the stability level (`beta`, `GA`) in the request. Instead, consumers simply request a version by date, and the paths within that version automatically resolve to their correct stability levels (`beta` or `GA`).
- Experimental stabilities are no longer supported; all paths are now either `beta` or `GA`.
- Changes to a path's stability are reflected by date, ensuring consistent access to `GA` or `beta` versions without ambiguity.

This results in a more straightforward API usage for clients, as they no longer need to be concerned with explicitly requesting stabilities, which reduces friction and simplifies interaction with the aggregated service APIs.

##### Example: Pre and Post Pivot Date Behavior

To better understand how simplified versioning works in comparison to the previous versioning model, let’s consider an example both before and after the pivot date of `2024-10-15`.

###### Pre Pivot Date Example (before 2024-10-15)

Consider a pet store API with services for `animals` and `petfood`. Assume the following versions:

- `animals` has:

  - `2024-10-01~beta`
  - `2024-10-01~experimental`

- `petfood` has:
  - `2024-09-20~GA`
  - `2024-10-01~beta`

If a client requested:

- **Request:**

  ```
  GET /api/2024-10-01~beta/animals
  ```

  - **Response:** This request would be served by the beta version of `/animals`.

- **Request:**

  ```
  GET /api/2024-10-01~experimental/animals
  ```

  - **Response:** This request would be served by the experimental version of `/animals`.

- **Request:**
  ```
  GET /api/2024-09-20~GA/petfood
  ```
  - **Response:** This request would be served by the GA version of `/petfood`.

Here, users must explicitly specify the stability (`~beta`, `~experimental`, or `~GA`), which adds complexity to the request.

###### Post Pivot Date Example (after 2024-10-15)

After the pivot date of `2024-10-15`, the simplified versioning model takes effect, where all paths are collated into a single spec with individual stabilities (`beta` or `GA`). Consider the following scenario:

- `animals` has:

  - Path `/animals` marked as `beta` on `2024-10-20`.

- `petfood` has:
  - Path `/petfood` marked as `GA` on `2024-10-20`.

If a client requested:

- **Request:**

  ```
  GET /api/2024-10-20/animals
  ```

  - **Response:** This request will be handled by the beta version of `/animals` without specifying the stability explicitly.

- **Request:**
  ```
  GET /api/2024-10-20/petfood
  ```
  - **Response:** This request will be handled by the GA version of `/petfood`.

With the simplified model, clients simply provide the date, and VU handles the rest, resolving the correct stability for each path automatically.

From 2024-10-15, Vervet Underground also integrates the simplified versioning model. With simplified versioning:

- API versions compiled by VU no longer require consumers to specify the stability level (`beta`, `GA`) in the request. Instead, consumers simply request a version by date, and the paths within that version automatically resolve to their correct stability levels (`beta` or `GA`).
- Experimental stabilities are no longer supported; all paths are now either `beta` or `GA`.
- Changes to a path's stability are reflected by date, ensuring consistent access to `GA` or `beta` versions without ambiguity.

This results in a more straightforward API usage for clients, as they no longer need to be concerned with explicitly requesting stabilities, which reduces friction and simplifies interaction with the aggregated service APIs.
