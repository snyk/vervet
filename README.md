# vervet

Vervet is a utility for managing versioned endpoints with OpenAPI v3.

# Motivation

Specifics of why and how we version endpoints is explained in better detail
[here](https://docs.google.com/document/d/1znmTSBIeQuAzAGEBeLHoQB5nW3_hucsSTmSSIm87biw/edit#heading=h.h25yqthzkhpj).

To summarize:

There are many opinions and approaches to how APIs should be versioned. We
often think of an API being versioned at the top-level, consisting of the
sum-total of all endpoints, the shape of their requests, responses,
conventions, types, and how they are processed.

This seems quite sensible, and indeed, consumers of an API need to depend on a
given contract defining what to expect.

However, development of such APIs works differently. As Snyk's API surface
increases with scale, different teams own different parts of the API, and it
will need to change at different rates. What these teams need are tools to
version their own section of the API independently, in such a way that
consumers of that API still get a cohesive API contract at all times, at the
point in time or stability level they need.

The right granularity for API consumers is a single version for an OpenAPI
contract defining the shape of all endpoints. This can be used to generate
client code, documentation, etc.

The right granularity for developers is a version per endpoint. When an
endpoint needs a breaking change, a team should be able to incubate and release
a new version without disrupting existing versions in use.

Vervet is a utility that bridges this gap between consumer and developer concerns.

## Installation

### NPM

    npm install -g @snyk/vervet

NPM packaging adapted from https://github.com/manifoldco/torus-cli.

### Source

Go >= 1.16 required.

    go build ./cmd/vervet

# Usage

Vervet compiles a single OpenAPI spec from a collection of OpenAPI specs, defined per endpoint per version.

See `vervet help` for a complete list and description of subcommands and options.

## Endpoint format

An endpoint is simply a directory structure of the form:

     endpoint/
     +- 2021-01-01
        +- spec.yaml
     +- 2021-06-21
        +- spec.yaml
     +- beta
        +- spec.yaml

where `endpoint` is the top-level path where the endpoint is defined (name is arbitrary). Its subdirectories are _versions_
each containing an OpenAPI v3 `spec.yaml`.

Each of these `spec.yaml` files must declare one and only one path, and they
all must be the same path (if they were different, they'd be different
endpoints!).

That's it! These directories can contain other files, generated code, etc.

## Versioning rules

When resolving an endpoint version, the most recent version less than or equal
to the requested version matches.

So, in the above example, resolving version "2021-03-31" would give you
"2021-01-01". Resolving "2021-06-24" gives you "2021-06-21".  And resolving
"2020-07-01" gives you nothing, because the endpoint had no version then.

There are two special version tags, "beta" and "experimental". These are always
newer than date-based versions (since we don't attempt to predict the future!),
and "experimental" is always newer than "beta".

## Compiling a versioned spec

Vervet processes endpoint spec directory structures described above into a
single consumer-facing OpenAPI spec. As an example:

    vervet compile -I common-includes.yaml /path/to/endpoints /path/to/output

will compile OpenAPI documents to `/path/to/output` at each distinct version
defined across all source endpoints found under `/path/to/endpoints`. The
optional `common-includes.yaml` is an OpenAPI document that is merged into each
compiled output document.

# Development

vervet uses a reference set of OpenAPI documents in `testdata/resources` in
tests. CLI tests compare runtime compiled output with pre-compiled, expected
output in `testdata/output` to detect regressions.

When introducing changes that intentionally change the content of compiled
output:

* Run `pre-commit.sh` to update the contents of `testdata/output`
* Verify that the compiled output is correct
* Commit the changes to `testdata/output` in your proposed branch

