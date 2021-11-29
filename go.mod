module github.com/snyk/vervet

go 1.16

require (
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20211112122917-428f8eabeeb3 // indirect
	github.com/bmatcuk/doublestar/v4 v4.0.2
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/frankban/quicktest v1.13.0
	github.com/getkin/kin-openapi v0.83.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.3.0
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/reflectwalk v1.0.2
	github.com/olekukonko/tablewriter v0.0.5
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/urfave/cli/v2 v2.3.0
	github.com/xanzy/ssh-agent v0.3.1 // indirect
	go.uber.org/multierr v1.7.0
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871 // indirect
	golang.org/x/net v0.0.0-20211118161319-6a13c67c3ce4 // indirect
	golang.org/x/sys v0.0.0-20211117180635-dee7805ff2e1 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

// Force git-go dependencies to use the latest go-crypto to resolve:
// - https://app.snyk.io/vuln/SNYK-GOLANG-GOLANGORGXCRYPTO-1083910
// - https://app.snyk.io/vuln/SNYK-GOLANG-GOLANGORGXCRYPTOSSH-551923
replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871
