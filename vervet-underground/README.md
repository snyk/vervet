# Problem

## API Program Axioms

1. The *Snyk API* will be composed of many *services*, each comprising parts of the whole product.
2. Each of these *services* contribute *resources* to that API.
3. Each *resource* within a service project may be developed, versioned and released independently at its own cadence.

[Vervet](https://github.com/snyk/vervet/) currently satisfies (b) and (c) by compiling a service-level versioned API aggregated from independently versioned resources in that service's project source.

(a) still needs to be addressed in order for the Snyk API and its versioning to be truly distributed across services.

## As below, so above

What's needed is a "Vervet" for microservices. Just as Vervet aggregates resource API specs "in the small" within a service, we need a "Vervet-of-Vervets" that can aggregate service API specs "in the large" across many services.

## Service-level API aggregation has different requirements

Aggregating APIs across services brings a unique and different set of challenges, raising questions:

- How are shared definitions in our API actually shared among services?
    - How are they versioned? How do they change over time?
    - What types are shared?
        - Resource data types do not need to be shared (other resources can link to them)
        - JSON API types (errors, links, relationships, top-level structure) or other patterns may emerge across resources in our APIs, that need to have a shared definition
        - Note that these types will not change very often, so a slow-moving semver versioning scheme is appropriate
- Are there linting rules that only make sense at the service level?
    - How do we ensure services are sunsetting versions according to policy?
- How are linting rules shared among services?
    - How are these versioned?
- How is validation and governance handled among services?

# Solution

Vervet Underground (VU) is a system that solves the above system-level API integration requirements.


## Minimum Viable

VU aggregates OpenAPI specs from multiple services and serves them up in a single place for:

- Docs
    - Docs will likely either render public `/openapi` directly client-side or periodically scrape & update
- Routing public v3 `/openapi` to VU's aggregated OpenAPI versions
- Add Akamai configuration to serve the blockstored specs initially for `/openapi`
    - Formal frontend presentation will be later

This unblocks docs for multi-service registry decomp.

### Details

- Could use block storage with a history of changes per service per version
- Static config that tells VU where to scrape upstream OpenAPI from services (registry and friends)
- Cron job to periodically scrape and update
    - Or we can try to make this push and set up a webhook...