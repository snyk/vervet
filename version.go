// Package vervet supports opinionated API versioning tools.
package vervet

import (
	"fmt"
	"strings"
	"time"
)

// Version defines an API version. API versions may be dates of the form
// "YYYY-mm-dd", or stability tags "beta", "experimental".
type Version struct {
	Date      time.Time
	Stability Stability
}

// DateString returns the string representation of the version date in
// YYYY-mm-dd form.
func (v *Version) DateString() string {
	return v.Date.Format("2006-01-02")
}

// String returns the string representation of the version in
// YYYY-mm-dd~Stability form.
func (v *Version) String() string {
	d := v.Date.Format("2006-01-02")
	if v.Stability != StabilityGA {
		return d + "~" + v.Stability.String()
	}
	return d
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
)

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
	panic(fmt.Sprintf("invalid stability value: %d", int(s)))
}

// ParseVersion parses a version string into a Version type, returning an error
// if the string is invalid.
func ParseVersion(s string) (*Version, error) {
	parts := strings.Split(s, "~")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid version %q", s)
	}
	d, err := time.ParseInLocation("2006-01-02", parts[0], time.UTC)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q", s)
	}
	stab := StabilityGA
	if len(parts) > 1 {
		stab, err = ParseStability(parts[1])
		if err != nil {
			return nil, err
		}
	}
	return &Version{Date: d.UTC(), Stability: stab}, nil
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
	default:
		return stabilityUndefined, fmt.Errorf("invalid stability %q", s)
	}
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
func (v *Version) Compare(vr *Version) int {
	dateCmp, stabilityCmp := v.compareDateStability(vr)
	if dateCmp != 0 {
		return dateCmp
	}
	return stabilityCmp
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

// VersionSlice is a sortable, searchable slice of Versions.
type VersionSlice []Version

// Resolve returns the most recent Version in the slice with equal or greater
// stability.
//
// This method requires that the VersionSlice has already been sorted with
// sort.Sort, otherwise behavior is undefined.
func (vs VersionSlice) Resolve(q Version) (*Version, error) {
	lower, curr, upper := 0, len(vs)/2, len(vs)
	if upper == 0 {
		// Nothing matches an empty slice.
		return nil, ErrNoMatchingVersion
	}
	for curr < upper && lower != upper-1 {
		dateCmp, stabilityCmp := vs[curr].compareDateStability(&q)
		if dateCmp > 0 {
			// Current version is more recent than the query, so it's our new
			// upper (exclusive) range limit to search.
			upper = curr
			curr = lower + (upper-lower)/2
		} else if dateCmp <= 0 {
			if stabilityCmp >= 0 {
				// Matching version found, so it's our new lower (inclusive)
				// range limit to search.
				lower = curr
			}
			// The edge is somewhere between here and the upper limit.
			curr = curr + (upper-curr)/2 + (upper-curr)%2
		}
	}
	// Did we find a match?
	dateCmp, stabilityCmp := vs[lower].compareDateStability(&q)
	if dateCmp <= 0 && stabilityCmp >= 0 {
		v := &vs[lower]
		return v, nil
	}
	return nil, ErrNoMatchingVersion
}

// Len implements sort.Interface.
func (vs VersionSlice) Len() int { return len(vs) }

// Less implements sort.Interface.
func (vs VersionSlice) Less(i, j int) bool {
	return vs[i].Compare(&vs[j]) < 0
}

// Swap implements sort.Interface.
func (vs VersionSlice) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }
