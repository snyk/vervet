package config_test

import (
	"reflect"
	"testing"

	"github.com/snyk/vervet/v8/config"
)

func TestOutput_ResolvePaths(t *testing.T) {
	tests := []struct {
		name    string
		subject *config.Output
		want    []string
	}{
		{
			name:    "nil",
			subject: nil,
			want:    []string{},
		},
		{
			name: "returns path if exists",
			subject: &config.Output{
				Path:  "path",
				Paths: []string{"path1", "path2"},
			},
			want: []string{"path"},
		},
		{
			name: "return paths if path is empty",
			subject: &config.Output{
				Path:  "",
				Paths: []string{"path1", "path2"},
			},
			want: []string{"path1", "path2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.subject.ResolvePaths(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolvePaths() = %v, want %v", got, tt.want)
			}
		})
	}
}
