package config_test

import (
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v8/config"
)

func TestOutput_deserialise(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		wantPaths   []string
		expectedErr string
	}{
		{
			name:        "nil",
			subject:     "{}",
			wantPaths:   []string(nil),
			expectedErr: "",
		},
		{
			name:        "returns path if exists",
			subject:     `{"path": "path1"}`,
			wantPaths:   []string{"path1"},
			expectedErr: "",
		},
		{
			name:        "return paths if exists",
			subject:     `{"paths": ["path1", "path2"]}`,
			wantPaths:   []string{"path1", "path2"},
			expectedErr: "",
		},
		{
			name:        "errors if both path and paths exist",
			subject:     `{"path": "path1", "paths": ["path1", "path2"]}`,
			wantPaths:   []string{},
			expectedErr: "output should specify one of 'path' or 'paths', not both",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			out := config.Output{}
			err := json.Unmarshal([]byte(tt.subject), &out)
			if tt.expectedErr != "" {
				c.Assert(err.Error(), qt.Equals, tt.expectedErr)
			} else {
				c.Assert(err, qt.IsNil)
				c.Assert(out.Paths, qt.DeepEquals, tt.wantPaths)
			}
		})
	}
}
