package scaffold

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

var manifestTests = []struct {
	desc     string
	version  string
	organize map[string]string
	err      string
}{{
	desc:     "invalid version",
	version:  "nope",
	organize: map[string]string{"foo": "foo"},
	err:      `unsupported manifest version "nope"`,
}, {
	desc:     "empty manifest - nil map",
	organize: map[string]string{},
	err:      `empty manifest`,
}, {
	desc:     "empty manifest - empty map",
	organize: map[string]string{},
	err:      `empty manifest`,
}, {
	desc:     "missing file",
	organize: map[string]string{"foo": "foo", "bar": "bar"},
	err:      `.*cannot stat source item ".*/bar":.*`,
}, {
	desc:     "ok, default version",
	organize: map[string]string{"foo": "foo"},
}, {
	desc:     "ok, version 1",
	version:  "1",
	organize: map[string]string{"foo": "foo"},
}}

func TestManifestValidate(t *testing.T) {
	c := qt.New(t)
	fakeSrc := c.Mkdir()
	c.Assert(ioutil.WriteFile(filepath.Join(fakeSrc, "foo"), []byte("foo"), 0666), qt.IsNil)
	for _, t := range manifestTests {
		m := &Manifest{
			Version:  t.version,
			Organize: t.organize,
		}
		if t.err == "" {
			c.Assert(m.validate(fakeSrc), qt.IsNil)
		} else {
			err := m.validate(fakeSrc)
			c.Assert(err, qt.ErrorMatches, t.err)
		}
	}
}
