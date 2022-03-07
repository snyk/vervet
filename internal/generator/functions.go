package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/dop251/goja"
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	builtinFuncs = template.FuncMap{
		"map": func(keyValues ...interface{}) (map[string]interface{}, error) {
			if len(keyValues)%2 != 0 {
				return nil, fmt.Errorf("invalid number of arguments to map")
			}
			m := make(map[string]interface{}, len(keyValues)/2)
			for i := 0; i < len(keyValues); i += 2 {
				k, ok := keyValues[i].(string)
				if !ok {
					return nil, fmt.Errorf("map keys must be strings")
				}
				m[k] = keyValues[i+1]
			}
			return m, nil
		},
		"indent": func(indent int, s string) string {
			return strings.ReplaceAll(s, "\n", "\n"+strings.Repeat(" ", indent))
		},
		"uncapitalize": func(s string) string {
			if len(s) > 1 {
				return strings.ToLower(s[0:1]) + s[1:]
			}
			return s
		},
		"capitalize": func(s string) string {
			if len(s) > 1 {
				return strings.ToUpper(s[0:1]) + s[1:]
			}
			return s
		},
		"replaceall":         strings.ReplaceAll,
		"pathOperations":     MapPathOperations,
		"resourceOperations": MapResourceOperations,
		"isOneOf": func(s *openapi3.Schema) bool {
			return s != nil && len(s.OneOf) > 0
		},
		"isAnyOf": func(s *openapi3.Schema) bool {
			return s != nil && len(s.AnyOf) > 0
		},
		"isAllOf": func(s *openapi3.Schema) bool {
			return s != nil && len(s.AllOf) > 0
		},
		"isAssociativeArray": func(s *openapi3.Schema) bool {
			return s != nil &&
				s.Type == "object" &&
				len(s.Properties) == 0 &&
				s.AdditionalPropertiesAllowed != nil &&
				*s.AdditionalPropertiesAllowed
		},
		"basename": filepath.Base,
	}
)

func withIncludeFunc(t *template.Template) *template.Template {
	return t.Funcs(template.FuncMap{
		"include": func(name string, data interface{}) (string, error) {
			buf := bytes.NewBuffer(nil)
			if err := t.ExecuteTemplate(buf, name, data); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
	})
}

var jsConsole = map[string]func(goja.FunctionCall) goja.Value{
	"log": func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i := range call.Arguments {
			args[i] = call.Arguments[i].Export()
		}
		log.Println(args...)
		return goja.Null()
	},
}

func (g *Generator) loadFunctions(filename string) error {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	vm := goja.New()
	_, err = vm.RunScript(filename, string(src))
	if err != nil {
		return err
	}
	module := vm.GlobalObject()
	if err != nil {
		return err
	}
	err = module.Set("console", jsConsole)
	if err != nil {
		return err
	}
	for _, key := range module.Keys() {
		fn, ok := goja.AssertFunction(module.Get(key))
		if !ok {
			// not a callable function
			continue
		}
		g.functions[key] = func(args ...interface{}) (interface{}, error) {
			jsArgs := make([]goja.Value, len(args))
			for i := range args {
				jsArgs[i] = vm.ToValue(args[i])
			}
			return fn(goja.Undefined(), jsArgs...)
		}
	}
	return nil
}
