package vervet

import (
	"fmt"
	"time"
)

// https://github.com/snyk/sweater-comb/blob/main/docs/principles/version.md
var VersionSunsetPolicy = map[Stability]time.Duration{
	StabilityGA:           180 * 24 * time.Hour,
	StabilityBeta:         90 * 24 * time.Hour,
	StabilityExperimental: 30 * 24 * time.Hour,
}

func GetSunsetDate(v Version) (time.Time, error) {
	duration, ok := VersionSunsetPolicy[v.Stability]
	if !ok {
		return time.Time{}, fmt.Errorf("unknown stability: %s", v.Stability)
	}
	return v.Date.Add(duration), nil
}
