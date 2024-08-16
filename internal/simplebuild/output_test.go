package simplebuild

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
)

func TestDocSet_WriteOutputs(t *testing.T) {
	c := qt.New(t)

	loader := openapi3.NewLoader()
	testDoc, err := loader.LoadFromData([]byte(minimalSpec))
	c.Assert(err, qt.IsNil)

	type args struct {
		cfg               config.Output
		appendOutputFiles bool
	}
	tests := []struct {
		name    string
		docs    DocSet
		args    args
		wantErr bool
		assert  func(*testing.T, args)
		setup   func(*testing.T, args)
	}{
		{
			name: "write the doc sets to outputs",
			args: args{
				cfg: config.Output{
					Path: t.TempDir(),
				},
			},
			docs: DocSet{
				{
					VersionDate: vervet.MustParseVersion("2024-01-01").Date,
					Doc:         testDoc,
				},
			},
			wantErr: false,
			assert: func(t *testing.T, args args) {
				t.Helper()
				files, err := filepath.Glob(filepath.Join(args.cfg.Path, "*"))
				c.Assert(err, qt.IsNil)
				c.Assert(files, qt.HasLen, 2)
				goEmbedContents, err := os.ReadFile(path.Join(args.cfg.Path, "embed.go"))
				c.Assert(err, qt.IsNil)
				c.Assert(string(goEmbedContents), qt.Contains, "2024-01-01")
			},
		},
		{
			name: "clears dir if appendOutputFiles is false",
			args: args{
				cfg: config.Output{
					Path: t.TempDir(),
				},
				appendOutputFiles: false,
			},
			docs: DocSet{
				{
					VersionDate: vervet.MustParseVersion("2024-01-01").Date,
					Doc:         testDoc,
				},
			},
			wantErr: false,
			setup: func(t *testing.T, args args) {
				t.Helper()
				err = os.WriteFile(path.Join(args.cfg.Path, "existing-file"), []byte("existing"), 0644)
				c.Assert(err, qt.IsNil)
			},
			assert: func(t *testing.T, args args) {
				t.Helper()
				files, err := filepath.Glob(filepath.Join(args.cfg.Path, "*"))
				c.Assert(err, qt.IsNil)
				c.Assert(files, qt.HasLen, 2)
				goEmbedContents, err := os.ReadFile(path.Join(args.cfg.Path, "embed.go"))
				c.Assert(err, qt.IsNil)
				c.Assert(string(goEmbedContents), qt.Contains, "2024-01-01")
			},
		},

		{
			name: "merges files if appendOutputFiles is true, embeds existing files",
			args: args{
				cfg: config.Output{
					Path: t.TempDir(),
				},
				appendOutputFiles: true,
			},
			docs: DocSet{
				{
					VersionDate: vervet.MustParseVersion("2024-01-01").Date,
					Doc:         testDoc,
				},
			},
			wantErr: false,
			setup: func(t *testing.T, args args) {
				t.Helper()
				err = os.WriteFile(path.Join(args.cfg.Path, "2024-02-01"), []byte("existing"), 0644)
				c.Assert(err, qt.IsNil)
				err = os.WriteFile(path.Join(args.cfg.Path, "2024-02-02"), []byte("existing"), 0644)
				c.Assert(err, qt.IsNil)
				err = os.WriteFile(path.Join(args.cfg.Path, "2024-02-03"), []byte("existing"), 0644)
				c.Assert(err, qt.IsNil)
			},
			assert: func(t *testing.T, args args) {
				t.Helper()
				files, err := filepath.Glob(filepath.Join(args.cfg.Path, "*"))
				c.Assert(err, qt.IsNil)
				c.Assert(files, qt.HasLen, 2+3)
				goEmbedContents, err := os.ReadFile(path.Join(args.cfg.Path, "embed.go"))
				c.Assert(err, qt.IsNil)
				c.Assert(string(goEmbedContents), qt.Contains, "2024-01-01")
				c.Assert(string(goEmbedContents), qt.Contains, "2024-02-01")
				c.Assert(string(goEmbedContents), qt.Contains, "2024-02-02")
				c.Assert(string(goEmbedContents), qt.Contains, "2024-02-03")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t, tt.args)
			}
			if err := tt.docs.WriteOutputs(tt.args.cfg, tt.args.appendOutputFiles); (err != nil) != tt.wantErr {
				t.Errorf("WriteOutputs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.assert != nil {
				tt.assert(t, tt.args)
			}
		})
	}
}

var minimalSpec = `---
paths: {}`
