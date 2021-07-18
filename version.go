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
		switch parts[1] {
		// wip endpoints get aggregated as experimental
		case "wip":
			stab = StabilityWIP
		case "experimental":
			stab = StabilityExperimental
		case "beta":
			stab = StabilityBeta
		default:
			return nil, fmt.Errorf("invalid version %q", s)
		}
	}
	return &Version{Date: d.UTC(), Stability: stab}, nil
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
	if v.Date.Before(vr.Date) {
		return -1
	}
	if v.Date.After(vr.Date) {
		return 1
	}
	// Dates are equal
	return 0 - v.Stability.Compare(vr.Stability)
}

// VersionDateStrings returns a slice of distinct version date strings for a
// slice of Versions. Consecutive duplicate dates are removed.
func VersionDateStrings(vs []*Version) []string {
	var result []string
	for i := range vs {
		ds := vs[i].DateString()
		if len(result) == 0 || result[len(result)-1] != ds {
			result = append(result, ds)
		}
	}
	return result
}
