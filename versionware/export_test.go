package versionware

import "time"

var DefaultValidatorConfig = defaultValidatorConfig

func (v *Validator) SetToday(today func() time.Time) {
	v.today = today
}
