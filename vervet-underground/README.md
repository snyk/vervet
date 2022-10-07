# Vervet Underground

# What Vervet Underground does and why

In order to understand _why_ Vervet Underground exists and the problem it solves, you should first become familiar with the [API versioning scheme](https://github.com/snyk/sweater-comb/blob/main/docs/principles/version.md) that Vervet supports. The main idea is, an API may be authored in parts, each of those parts may be versioned, and all the distinct versions are assembled to produce a cohesive timeline of versions for the entire service.

Just as Vervet compiles a timeline of OpenAPI versions for a single service from independently versioned parts, Vervet Underground (VU) compiles a timeline of OpenAPI spec versions for a SaaS from independently versioned microservices, each of which contributes parts of the SaaS API.

# Service API aggregation by example

## The pet store gets microservices

To illustrate how this works in practice, let's deconstruct a pet store into two services:

* `petfood.default.svc.cluster.local`, which knows about pet food.
* `animals.default.svc.cluster.local`, which knows about animals.

For sake of example, let's assume the following versions are published by each service:

petfood has:
- `2021-07-04~experimental`
- `2021-08-09~beta`
- `2021-08-09~experimental` (beta released, with some parts still experimental, so both are published)
- `2021-09-14` .            (first GA)
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
