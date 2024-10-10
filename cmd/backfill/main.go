package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/snyk/cerberus/v2/openapi3cerb"
	"github.com/snyk/rest-go-libs/v5/authz"
	"golang.org/x/exp/maps"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/internal/simplebuild"
)

type Endpoint struct {
	Path         *regexp.Regexp
	PathStr      string
	Method       string
	Permissions  map[authz.Permission]interface{}
	Entitlements map[authz.Entitlement]interface{}
}

func main() {
	ctx := context.Background()
	fmt.Println("backfill data")
	endpoints, err := getEndpoints("endpoints.csv")
	if err != nil {
		panic(err)
	}
	fmt.Println("loaded endpoints from csv")

	project, err := getProject()
	if err != nil {
		panic(err)
	}
	for _, apiConfig := range project.APIs {
		ops, err := simplebuild.LoadPaths(ctx, apiConfig)
		if err != nil {
			panic(err)
		}
		fmt.Printf("loaded %d operations from %s\n", len(ops), apiConfig.Name)
		for opKey, opVersions := range ops {
			apiName := apiConfig.Name
			if apiName == "registry-internal-rpc-api" {
				// internal rcp uses the same prefixes as internal
				apiName = "registry-internal-api"
			}
			csvOp := findMatch(opKey.Path, opKey.Method, endpoints[apiName])
			if csvOp == nil {
				fmt.Println(" ! missing entry from csv for", opKey.Method, opKey.Path)
				continue
			}

			for _, op := range opVersions {
				cerbBlock := &openapi3cerb.Extension{}
				err := openapi3cerb.ExtractExtension(op.Operation.Extensions, &cerbBlock)
				if err != nil {
					panic(err)
				}
				matching(cerbBlock, *csvOp, opKey, op.Version)
			}
		}
	}
}

func matching(extension *openapi3cerb.Extension, endpoint Endpoint, opKey simplebuild.OpKey, version vervet.Version) {
	entA := extension.Authorization.Resource.Entitlements
	entB := maps.Keys(endpoint.Entitlements)
	slices.Sort(entA)
	slices.Sort(entB)
	if len(entA) != 0 && len(entB) != 0 && !reflect.DeepEqual(entA, entB) {
		fmt.Printf(
			" ! %s %s (%s) entitlement mismatch\n    - spec: %s\n    - csv: %s\n",
			opKey.Method,
			opKey.Path,
			version.String(),
			entA,
			entB,
		)
	}

	permA := extension.Authorization.Resource.Permissions
	// permissions in openapi docs have the resource implicitly applied, eg
	// org.project.read => project.read
	for idx, perm := range permA {
		permA[idx] = authz.Permission(
			fmt.Sprintf("%s.%s", extension.Authorization.Resource.Type, perm),
		)
	}
	permB := maps.Keys(endpoint.Permissions)
	slices.Sort(permA)
	slices.Sort(permB)
	if len(permA) != 0 && len(permB) != 0 && !reflect.DeepEqual(permA, permB) {
		fmt.Printf(
			" ! %s %s (%s) permissions mismatch\n    - spec: %s\n    - csv: %s\n",
			opKey.Method,
			opKey.Path,
			version.String(),
			permA,
			permB,
		)
	}
}

func parsePrefix(path string) (string, string) {
	mapping := map[string]string{
		"/api/v3":       "registry-v3-api",
		"/api/hidden":   "registry-hidden-api",
		"/api/internal": "registry-internal-api",
		"/api/sarif":    "sarif",
	}

	sectionMap := map[string]string{
		"orgPublicId":            "org_id",
		"orgId":                  "org_id",
		"groupPublicId":          "group_id",
		"groupId":                "group_id",
		"policyPublicId":         "policy_id",
		"serviceAccountPublicId": "serviceaccount_id",
		"ssoPublicId":            "sso_id",
		"projectId":              "project_id",
		"membershipPublicId":     "membership_id",
		"userPublicId":           "user_id",
		"invitationPublicId":     "invite_id",
		"ruleId":                 "rule_id",
	}

	specialCases := map[string][]string{
		"/api/internal/activity_implementor/kubernetes_integration/authorize":         {"/kubernetes_integration/authorize", "registry-internal-api"},
		"/api/internal/activity_implementor/license_policy/get_by_project_attributes": {"/license_policy/get_by_project_attributes", "registry-internal-api"},
		"/api/internal/activity_implementor/recurring_test/iac":                       {"/recurring_test/iac", "registry-internal-api"},
		"/api/internal/activity_implementor/recurring_test/sast":                      {"/recurring_test/sast", "registry-internal-api"},
		"/api/internal/activity_implementor/security_policy/apply":                    {"/security_policy/apply", "registry-internal-api"},
		"/api/internal/rpc/test_usage/opensource/cli/check":                           {"/rpc/opensource/cli/check", "registry-internal-api"},
		"/api/internal/rpc/test_usage/opensource/track":                               {"/rpc/opensource/track", "registry-internal-api"},

		"/api/sarif/": {"/sarif", "sarif"},

		"/api/v3/orgs/:orgPublicId/issues/detail/code/:id": {"/orgs/{org_id}/issues/detail/code/{issue_id}", "registry-v3-api"},
		"/api/v3/orgs/:orgPublicId/code_issue_details/:id": {"/orgs/{org_id}/code_issue_details/{issue_id}", "registry-v3-api"},
	}

	if cases, ok := specialCases[path]; ok {
		return cases[0], cases[1]
	}

	for key, val := range mapping {
		if !strings.HasPrefix(path, key) {
			continue
		}

		trimmed, _ := strings.CutPrefix(path, key)
		sections := strings.Split(trimmed, "/")
		for idx, section := range sections {
			if strings.HasPrefix(section, ":") {
				var replacement string
				if sub, ok := sectionMap[section[1:]]; ok {
					replacement = sub
				} else {
					replacement = section[1:]
				}
				sections[idx] = replacement
			}
		}
		replaced := strings.Join(sections, "/")
		return replaced, val
	}

	return "", ""
}

func getEndpoints(csvPath string) (map[string][]Endpoint, error) {
	endpoints := make(map[string][]Endpoint)
	file, err := os.Open(csvPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	// First row is header
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}
	for {
		data, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return nil, err
			}
		}

		path, apiName := parsePrefix(data[1])
		if path == "" {
			// not an api we care about
			continue
		}
		method := data[0]

		var apiEndpoints []Endpoint
		if entry, ok := endpoints[apiName]; ok {
			apiEndpoints = entry
		} else {
			apiEndpoints = []Endpoint{}
		}

		ep := findByStr(path, method, apiEndpoints)
		if ep == nil {
			ep = &Endpoint{
				Path:         buildRegex(path),
				PathStr:      path,
				Method:       method,
				Permissions:  make(map[authz.Permission]interface{}),
				Entitlements: make(map[authz.Entitlement]interface{}),
			}
			apiEndpoints = append(apiEndpoints, *ep)
		}

		if data[5] != "" {
			for _, permission := range strings.Split(data[5], ";") {
				ep.Permissions[authz.Permission(permission)] = struct{}{}
			}
		}
		if data[6] != "" {
			for _, entitlement := range strings.Split(data[6], ";") {
				ep.Entitlements[authz.Entitlement(entitlement)] = struct{}{}
			}
		}

		endpoints[apiName] = apiEndpoints
	}
	return endpoints, nil
}

func findByStr(path, method string, haystack []Endpoint) *Endpoint {
	for _, entry := range haystack {
		if entry.PathStr == path && entry.Method == method {
			return &entry
		}
	}
	return nil
}

func findMatch(path, method string, haystack []Endpoint) *Endpoint {
	for _, entry := range haystack {
		if entry.Method == method && entry.Path.MatchString(path) {
			return &entry
		}
	}
	return nil
}

func getProject() (*config.Project, error) {
	var project *config.Project
	configPath := ".vervet.yaml"
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", configPath, err)
	}
	defer f.Close()
	project, err = config.Load(f)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func buildRegex(path string) *regexp.Regexp {
	segments := []string{}
	current := ""
	parens := 0
	for _, ch := range path[1:] {
		if ch == '/' && parens == 0 {
			segments = append(segments, current)
			current = ""
			continue
		}
		if ch == '(' {
			parens += 1
		} else if ch == ')' {
			parens -= 1
		}
		current += string(ch)
	}
	segments = append(segments, current)

	res := ""
	for _, segment := range segments {
		raw, found := strings.CutPrefix(segment, "RegExp(/^")
		if found {
			res += strings.TrimSuffix(raw, "$/i)")
		} else {
			res += "\\/" + raw
		}
	}
	re := fmt.Sprintf("^%s$", res)
	return regexp.MustCompile(re)
}
