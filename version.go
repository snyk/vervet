// Package vervet supports opinionated API versioning tools.
package vervet

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

var timeNow = time.Now

// Version defines an API version. API versions may be dates of the form
// "YYYY-mm-dd", or stability tags "beta", "experimental".
type Version struct {
	Date      time.Time
	Stability Stability
}

// DateString returns the string representation of the version date in
// YYYY-mm-dd form.
func (v Version) DateString() string {
	return v.Date.Format("2006-01-02")
}

// String returns the string representation of the version in
// YYYY-mm-dd~Stability form. This method will panic if the value is empty.
func (v Version) String() string {
	d := v.Date.Format("2006-01-02")
	if v.Stability != StabilityGA {
		return d + "~" + v.Stability.String()
	}
	return d
}

// AddDays returns the version corresponding to adding the given number of days
// to the version date.
func (v Version) AddDays(days int) Version {
	return Version{
		Date:      v.Date.AddDate(0, 0, days),
		Stability: v.Stability,
	}
}

// Stability defines the stability level of the version.
type Stability int

const (
	stabilityUndefined Stability = iota

	// StabilityWIP means the API is a work-in-progress and not yet ready.
	StabilityWIP Stability = iota

	// StabilityExperimental means the API is experimental and still subject to
	// drastic change.
	StabilityExperimental Stability = iota

	// StabilityBeta means the API is becoming more stable, but may undergo some
	// final changes before being released.
	StabilityBeta Stability = iota

	// StabilityGA means the API has been released and will not change.
	StabilityGA Stability = iota

	numStabilityLevels = iota
)

// String returns a string representation of the stability level. This method
// will panic if the value is empty.
func (s Stability) String() string {
	switch s {
	case StabilityWIP:
		return "wip"
	case StabilityExperimental:
		return "experimental"
	case StabilityBeta:
		return "beta"
	case StabilityGA:
		return "ga"
	}
	panic(fmt.Sprintf("invalid stability (%d)", int(s)))
}

// ParseVersion parses a version string into a Version type, returning an error
// if the string is invalid.
func ParseVersion(s string) (Version, error) {
	parts := strings.Split(s, "~")
	if len(parts) < 1 {
		return Version{}, fmt.Errorf("invalid version %q", s)
	}
	d, err := time.ParseInLocation("2006-01-02", parts[0], time.UTC)
	if err != nil {
		return Version{}, fmt.Errorf("invalid version %q", s)
	}
	stab := StabilityGA
	if len(parts) > 1 {
		stab, err = ParseStability(parts[1])
		if err != nil {
			return Version{}, err
		}
	}
	return Version{Date: d.UTC(), Stability: stab}, nil
}

// MustParseVersion parses a version string into a Version type, panicking if
// the string is invalid.
func MustParseVersion(s string) Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

// ParseStability parses a stability string into a Stability type, returning an
// error if the string is invalid.
func ParseStability(s string) (Stability, error) {
	switch s {
	case "wip":
		return StabilityWIP, nil
	case "experimental":
		return StabilityExperimental, nil
	case "beta":
		return StabilityBeta, nil
	case "ga":
		return StabilityGA, nil
	default:
		return stabilityUndefined, fmt.Errorf("invalid stability %q", s)
	}
}

// MustParseStability parses a stability string into a Stability type,
// panicking if the string is invalid.
func MustParseStability(s string) Stability {
	stab, err := ParseStability(s)
	if err != nil {
		panic(err)
	}
	return stab
}

// Compare returns -1 if the given stability level is less than, 0 if equal to,
// and 1 if greater than the caller target stability level.
func (s Stability) Compare(sr Stability) int {
	if s < sr {
		return -1
	} else if s > sr {
		return 1
	}
	return 0
}

// Compare returns -1 if the given version is less than, 0 if equal to, and 1
// if greater than the caller target version.
func (v Version) Compare(vr Version) int {
	dateCmp, stabilityCmp := v.compareDateStability(&vr)
	if dateCmp != 0 {
		return dateCmp
	}
	return stabilityCmp
}

// DeprecatedBy returns true if the given version deprecates the caller target
// version.
func (v Version) DeprecatedBy(vr Version) bool {
	dateCmp, stabilityCmp := v.compareDateStability(&vr)
	// A version is deprecated by a newer version of equal or greater stability.
	return dateCmp == -1 && stabilityCmp <= 0
}

const (
	// SunsetWIP is the duration past deprecation after which a work-in-progress version may be sunset.
	SunsetWIP = 0

	// SunsetExperimental is the duration past deprecation after which an experimental version may be sunset.
	SunsetExperimental = 24 * time.Hour

	// SunsetBeta is the duration past deprecation after which a beta version may be sunset.
	SunsetBeta = 91 * 24 * time.Hour

	// SunsetGA is the duration past deprecation after which a GA version may be sunset.
	SunsetGA = 181 * 24 * time.Hour
)

// Sunset returns, given a potentially deprecating version, the eligible sunset
// date and whether the caller target version would actually be deprecated and
// sunset by the given version.
func (v Version) Sunset(vr Version) (time.Time, bool) {
	if !v.DeprecatedBy(vr) {
		return time.Time{}, false
	}
	switch v.Stability {
	case StabilityWIP:
		return vr.Date.Add(SunsetWIP), true
	case StabilityExperimental:
		return vr.Date.Add(SunsetExperimental), true
	case StabilityBeta:
		return vr.Date.Add(SunsetBeta), true
	case StabilityGA:
		return vr.Date.Add(SunsetGA), true
	}
	return time.Time{}, false
}

// compareDateStability returns the comparison of both the date and stability
// between two versions. Used internally where these need to be evaluated
// independently, such as when searching for the best matching version.
func (v *Version) compareDateStability(vr *Version) (int, int) {
	dateCmp := 0
	if v.Date.Before(vr.Date) {
		dateCmp = -1
	} else if v.Date.After(vr.Date) {
		dateCmp = 1
	}
	stabilityCmp := v.Stability.Compare(vr.Stability)
	return dateCmp, stabilityCmp
}

// VersionDateStrings returns a slice of distinct version date strings for a
// slice of Versions. Consecutive duplicate dates are removed.
func VersionDateStrings(vs []Version) []string {
	var result []string
	for i := range vs {
		ds := vs[i].DateString()
		if len(result) == 0 || result[len(result)-1] != ds {
			result = append(result, ds)
		}
	}
	return result
}

// VersionSlice is a sortable slice of Versions.
type VersionSlice []Version

// VersionIndex provides a search over versions, resolving which version is in
// effect for a given date and stability level.
type VersionIndex struct {
	versions []effectiveVersion
}

type effectiveVersion struct {
	date        time.Time
	stabilities [numStabilityLevels]time.Time
}

// NewVersionIndex returns a new VersionIndex of the given versions. The given
// VersionSlice will be sorted.
func NewVersionIndex(vs VersionSlice) (vi VersionIndex) {
	sort.Sort(vs)
	evIndex := -1
	currentStabilities := [numStabilityLevels]time.Time{}
	for i := range vs {
		if evIndex == -1 || !vi.versions[evIndex].date.Equal(vs[i].Date) {
			vi.versions = append(vi.versions, effectiveVersion{
				date:        vs[i].Date,
				stabilities: [numStabilityLevels]time.Time{},
			})
			evIndex++
			for stab, date := range currentStabilities {
				vi.versions[evIndex].stabilities[stab] = date
			}
		}
		vi.versions[evIndex].stabilities[vs[i].Stability] = vs[i].Date
		currentStabilities[vs[i].Stability] = vs[i].Date
	}
	return vi
}

// resolveIndex performs a binary search on the stability versions in effect on
// the query date.
func (vi *VersionIndex) resolveIndex(query time.Time) (int, error) {
	if len(vi.versions) == 0 || vi.versions[0].date.After(query) {
		return -1, ErrNoMatchingVersion
	}
	lower, curr, upper := 0, len(vi.versions)/2, len(vi.versions)
	for lower < upper-1 {
		if vi.versions[curr].date.After(query) {
			upper = curr
		} else {
			lower = curr
		}
		curr = lower + (upper-lower)/2
	}
	return lower, nil
}

// Resolve returns the released version effective on the query version date at
// the given version stability. Returns ErrNoMatchingVersion if no version matches.
//
// Resolve should be used on a collection of already "compiled" or
// "collated" API versions.
func (vi *VersionIndex) Resolve(query Version) (Version, error) {
	i, err := vi.resolveIndex(query.Date)
	if err != nil {
		return Version{}, err
	}
	for stab := query.Stability; stab < numStabilityLevels; stab++ {
		if stabDate := vi.versions[i].stabilities[stab]; !stabDate.IsZero() {
			return Version{Date: stabDate, Stability: stab}, nil
		}
	}
	return Version{}, ErrNoMatchingVersion
}

// resolveForBuild returns the most stable version effective on the query
// version date with respect to the given version stability. Returns
// ErrNoMatchingVersion if no version matches.
//
// Use resolveForBuild when resolving version deprecation and effective releases
// _within a single resource_ during the "compilation" or "collation" process.
func (vi *VersionIndex) resolveForBuild(query Version) (Version, error) {
	i, err := vi.resolveIndex(query.Date)
	if err != nil {
		return Version{}, err
	}
	var matchDate time.Time
	var matchStab Stability
	for stab := query.Stability; stab < numStabilityLevels; stab++ {
		if stabDate := vi.versions[i].stabilities[stab]; !stabDate.IsZero() && !stabDate.Before(matchDate) && !stabDate.After(query.Date) {
			matchDate, matchStab = stabDate, stab
		}
	}
	if matchDate.IsZero() {
		return Version{}, ErrNoMatchingVersion
	}
	return Version{Date: matchDate, Stability: matchStab}, nil
}

// Deprecates returns the version that deprecates the given version in the
// slice.
func (vi *VersionIndex) Deprecates(q Version) (Version, bool) {
	match, err := vi.resolveIndex(q.Date)
	if err == ErrNoMatchingVersion {
		return Version{}, false
	}
	if err != nil {
		panic(err)
	}
	for i := match + 1; i < len(vi.versions); i++ {
		for stab := q.Stability; stab < numStabilityLevels; stab++ {
			if stabDate := vi.versions[i].stabilities[stab]; stabDate.After(q.Date) {
				return Version{
					Date:      vi.versions[i].date,
					Stability: stab,
				}, true
			}
		}
	}
	return Version{}, false
}

// Len implements sort.Interface.
func (vs VersionSlice) Len() int { return len(vs) }

// Less implements sort.Interface.
func (vs VersionSlice) Less(i, j int) bool {
	return vs[i].Compare(vs[j]) < 0
}

// Swap implements sort.Interface.
func (vs VersionSlice) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }

// Strings returns a slice of string versions
func (vs VersionSlice) Strings() []string {
	s := make([]string, len(vs))
	for i := range vs {
		s[i] = vs[i].String()
	}
	return s
}

// Lifecycle defines the release lifecycle.
type Lifecycle int

const (
	lifecycleUndefined Lifecycle = iota

	// LifecycleUnreleased means the version has not been released yet.
	LifecycleUnreleased Lifecycle = iota

	// LifecycleReleased means the version is released.
	LifecycleReleased Lifecycle = iota

	// LifecycleDeprecated means the version is deprecated.
	LifecycleDeprecated Lifecycle = iota

	// LifecycleSunset means the version is eligible to be sunset.
	LifecycleSunset Lifecycle = iota

	// ExperimentalTTL is the duration after which experimental releases expire
	// and should be considered sunset.
	ExperimentalTTL = 90 * 24 * time.Hour
)

// ParseLifecycle parses a lifecycle string into a Lifecycle type, returning an
// error if the string is invalid.
func ParseLifecycle(s string) (Lifecycle, error) {
	switch s {
	case "released":
		return LifecycleReleased, nil
	case "deprecated":
		return LifecycleDeprecated, nil
	case "sunset":
		return LifecycleSunset, nil
	default:
		return lifecycleUndefined, fmt.Errorf("invalid lifecycle %q", s)
	}
}

// String returns a string representation of the lifecycle stage. This method
// will panic if the value is empty.
func (l Lifecycle) String() string {
	switch l {
	case LifecycleReleased:
		return "released"
	case LifecycleDeprecated:
		return "deprecated"
	case LifecycleSunset:
		return "sunset"
	}
	panic(fmt.Sprintf("invalid lifecycle (%d)", int(l)))
}

func (l Lifecycle) Valid() bool {
	switch l {
	case LifecycleReleased, LifecycleDeprecated, LifecycleSunset:
		return true
	default:
		return false
	}
}

// LifecycleAt returns the Lifecycle of the version at the given time. If the
// time is the zero value (time.Time{}), then the following are used to
// determine the reference time:
//
// If VERVET_LIFECYCLE_AT is set to an ISO date string of the form YYYY-mm-dd,
// this date is used as the reference time for deprecation, at midnight UTC.
//
// Otherwise `time.Now().UTC()` is used for the reference time.
//
// The current time is always used for determining whether a version is unreleased.
func (v *Version) LifecycleAt(t time.Time) Lifecycle {
	if t.IsZero() {
		t = defaultLifecycleAt()
	}
	deprecationDelta := t.Sub(v.Date)
	releaseDelta := timeNow().UTC().Sub(v.Date)
	if releaseDelta < 0 {
		return LifecycleUnreleased
	}
	if v.Stability.Compare(StabilityExperimental) <= 0 {
		if v.Stability == StabilityWIP {
			return LifecycleSunset
		}
		// experimental
		if deprecationDelta > ExperimentalTTL {
			return LifecycleSunset
		}
		return LifecycleDeprecated
	}
	return LifecycleReleased
}

func defaultLifecycleAt() time.Time {
	if dateStr := os.Getenv("VERVET_LIFECYCLE_AT"); dateStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC); err == nil {
			return t
		}
	}
	return timeNow().UTC()
}
