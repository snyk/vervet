package config

import "fmt"

const (
	defaultSweaterCombImage = "gcr.io/snyk-main/sweater-comb:latest"
	defaultOpticCIImage     = "snyk/sweater-comb:latest"
)

var defaultSpectralExtraArgs = []string{"--format", "text"}

// Linters defines a named map of Linter instances.
type Linters map[string]*Linter

// Linter describes a set of standards and rules that an API should satisfy.
type Linter struct {
	Name        string             `json:"-"`
	Description string             `json:"description,omitempty"`
	Spectral    *SpectralLinter    `json:"spectral"`
	SweaterComb *SweaterCombLinter `json:"sweater-comb"`
	OpticCI     *OpticCILinter     `json:"optic-ci"`
}

func (l *Linter) validate() error {
	nlinters := 0
	if l.Spectral != nil {
		nlinters++
	}
	if l.SweaterComb != nil {
		nlinters++
	}
	if l.OpticCI != nil {
		nlinters++
	}
	switch nlinters {
	case 0:
		return fmt.Errorf("missing configuration (linters.%s)", l.Name)
	case 1:
		return nil
	default:
		return fmt.Errorf("a linter may only be of one type (linters.%s)", l.Name)
	}
}

// SpectralLinter identifies a Linter as a collection of Spectral rulesets.
type SpectralLinter struct {

	// Rules are a list of Spectral ruleset file locations
	Rules []string `json:"rules"`

	// ExtraArgs may be used to pass extra arguments to `spectral lint`. If not
	// specified, the default arguments `--format text` are used when running
	// spectral. The `-r` flag must not be specified here, as this argument is
	// automatically added from the Rules setting above.
	//
	// See https://meta.stoplight.io/docs/spectral/ZG9jOjI1MTg1-spectral-cli
	// for the options supported.
	ExtraArgs []string `json:"extraArgs"`
}

// SweaterCombLinter identifies a Sweater Comb Linter, which is distributed as
// a self-contained docker image.
type SweaterCombLinter struct {
	// Image identifies the Sweater Comb docker image to use for linting.
	Image string

	// Rules are a list of Spectral ruleset file locations
	// These may be absolute paths to Sweater Comb rules, such as /rules/apinext.yaml.
	// Or, they may be relative paths to files in this project.
	Rules []string `json:"rules"`

	// ExtraArgs may be used to pass extra arguments to `spectral lint`. The
	// Sweater Comb image includes Spectral. This has the same function as
	// SpectralLinter.ExtraArgs above.
	ExtraArgs []string `json:"extraArgs"`
}

// OpticCILinter identifies an Optic CI Linter, which is distributed as
// a self-contained docker image.
type OpticCILinter struct {
	// Image identifies the Optic CI docker image to use for linting.
	Image string

	// Script identifies the path to the Optic CI script to use for linting.
	// Mutually exclusive with Image; if Script is specified Docker will not be
	// used.
	Script string

	// Original is where to source the original version of an OpenAPI spec file
	// when comparing. If empty, all changes are assumed to be new additions.
	Original string `json:"original,omitempty"`

	// Proposed is where to source the proposed changed version of an OpenAPI
	// spec file when comparing. If empty, this is assumed to be the
	// local working copy.
	Proposed string `json:"proposed,omitempty"`

	// Debug turns on debug logging.
	Debug bool `json:"debug,omitempty"`

	// CIContext specifies the location of an optional CI context JSON file to
	// use when performing an optic-ci compare or bulk-compare. If this file is
	// present when vervet lints, it will be passed to optic using the
	// --ci-context option.
	//
	// See https://www.npmjs.com/package/@useoptic/api-checks#expected-contexts
	// for more information.
	CIContext string `json:"ciContext"`

	// UploadResults indicates that the comparison results from optic-ci should
	// be uploaded to Optic Cloud. This only happens if the file specified by
	// CIContext exists.
	UploadResults bool `json:"uploadResults"`

	// Exceptions are files that are excluded from CI checks. This is an escape
	// hatch of last resort, if a file needs to land and can't pass CI yet.
	// They are specified as a mapping from project relative path to sha256
	// sums of that spec file that is exempt. This makes the exception very
	// narrow -- only a specific version of a specific file is skipped, after
	// outside review and approval.
	Exceptions map[string][]string

	// ExtraArgs may be used to pass extra arguments to `optic-ci`.
	ExtraArgs []string `json:"extraArgs"`
}

func (l Linters) init() error {
	for name, linter := range l {
		if linter == nil {
			return fmt.Errorf("missing linter definition (linters.%s)", name)
		}
		linter.Name = name
		if err := linter.validate(); err != nil {
			return err
		}
		if linter.Spectral != nil && len(linter.Spectral.ExtraArgs) == 0 {
			linter.Spectral.ExtraArgs = defaultSpectralExtraArgs
		}
		if linter.SweaterComb != nil {
			if len(linter.SweaterComb.ExtraArgs) == 0 {
				linter.SweaterComb.ExtraArgs = defaultSpectralExtraArgs
			}
			if linter.SweaterComb.Image == "" {
				linter.SweaterComb.Image = defaultSweaterCombImage
			}
		}
		if linter.OpticCI != nil {
			if linter.OpticCI.Image == "" && linter.OpticCI.Script == "" {
				linter.OpticCI.Image = defaultOpticCIImage
			}
		}
	}
	return nil
}
