package vervet

// TimeNow is a patchable func pointer to obtain time.Now in the
// version.go file, used for mocking time in tests.
var TimeNow = &timeNow
