package optic

// Context provides Optic with external information needed in order to process
// API versioning lifecycle rules. For example, lifecycle rules need to know
// when a change is occurring, and what other versions have deprecated the
// OpenAPI spec version being evaluated.
type Context struct {
	// ChangeDate is when the proposed change would occur.
	ChangeDate string `json:"changeDate"`

	// ChangeResource is the proposed change resource name.
	ChangeResource string `json:"changeResource"`

	// ChangeVersion is the proposed change version.
	ChangeVersion Version `json:"changeVersion"`

	// ResourceVersions describes other resource version releases.
	ResourceVersions ResourceVersionReleases `json:"resourceVersions,omitempty"`
}

// Version describes an API resource version, a date and a stability.
// Stability is assumed to be GA if not specified.
type Version struct {
	Date      string `json:"date"`
	Stability string `json:"stability,omitempty"`
}

// ResourceVersionReleases describes resource version releases.
type ResourceVersionReleases map[string]VersionStabilityReleases

// VersionStabilityReleases describes version releases.
type VersionStabilityReleases map[string]StabilityReleases

// StabilityReleases describes stability releases.
type StabilityReleases map[string]Release

// Release describes a single resource-version-stability release.
type Release struct {
	// DeprecatedBy indicates the other release version that deprecates this
	// release.
	DeprecatedBy Version `json:"deprecatedBy"`
}
