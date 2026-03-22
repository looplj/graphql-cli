package schema

import (
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type FindScope struct {
	Query    bool
	Mutation bool
	Type     bool
	Input    bool
	Enum     bool
}

func (s FindScope) IsEmpty() bool {
	return !s.Query && !s.Mutation && !s.Type && !s.Input && !s.Enum
}

type FindResult struct {
	Kind       string // "query", "mutation", "type", "input", "enum", "interface", "union", "scalar"
	Name       string
	Definition string
}

func ParseAndFind(sdl string, keyword string, scope FindScope) ([]FindResult, error) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  "schema",
		Input: sdl,
	})
	if err != nil {
		return nil, err
	}

	searchAll := scope.IsEmpty()
	keyword = strings.ToLower(keyword)

	var results []FindResult

	if searchAll || scope.Query {
		if schema.Query != nil {
			for _, f := range schema.Query.Fields {
				if isBuiltinField(f.Name) {
					continue
				}

				if keyword == "" || strings.Contains(strings.ToLower(f.Name), keyword) {
					results = append(results, FindResult{
						Kind:       "query",
						Name:       f.Name,
						Definition: formatField(f),
					})
				}
			}
		}
	}

	if searchAll || scope.Mutation {
		if schema.Mutation != nil {
			for _, f := range schema.Mutation.Fields {
				if isBuiltinField(f.Name) {
					continue
				}

				if keyword == "" || strings.Contains(strings.ToLower(f.Name), keyword) {
					results = append(results, FindResult{
						Kind:       "mutation",
						Name:       f.Name,
						Definition: formatField(f),
					})
				}
			}
		}
	}

	if searchAll || scope.Type || scope.Input || scope.Enum {
		for _, t := range schema.Types {
			if isBuiltinType(t.Name) {
				continue
			}

			if keyword != "" && !strings.Contains(strings.ToLower(t.Name), keyword) {
				continue
			}

			switch t.Kind {
			case ast.Object:
				if !searchAll && !scope.Type {
					continue
				}

				if t.Name == "Query" || t.Name == "Mutation" || t.Name == "Subscription" {
					continue
				}

				results = append(results, FindResult{
					Kind:       "type",
					Name:       t.Name,
					Definition: formatTypeDef(t),
				})
			case ast.InputObject:
				if !searchAll && !scope.Input {
					continue
				}

				results = append(results, FindResult{
					Kind:       "input",
					Name:       t.Name,
					Definition: formatTypeDef(t),
				})
			case ast.Enum:
				if !searchAll && !scope.Enum {
					continue
				}

				results = append(results, FindResult{
					Kind:       "enum",
					Name:       t.Name,
					Definition: formatTypeDef(t),
				})
			case ast.Interface:
				if !searchAll && !scope.Type {
					continue
				}

				results = append(results, FindResult{
					Kind:       "interface",
					Name:       t.Name,
					Definition: formatTypeDef(t),
				})
			case ast.Union:
				if !searchAll && !scope.Type {
					continue
				}

				results = append(results, FindResult{
					Kind:       "union",
					Name:       t.Name,
					Definition: formatTypeDef(t),
				})
			case ast.Scalar:
				if !searchAll && !scope.Type {
					continue
				}

				if isBuiltinScalar(t.Name) {
					continue
				}

				results = append(results, FindResult{
					Kind:       "scalar",
					Name:       t.Name,
					Definition: formatTypeDef(t),
				})
			}
		}
	}

	return results, nil
}

func isBuiltinField(name string) bool {
	return strings.HasPrefix(name, "__")
}

func isBuiltinType(name string) bool {
	if strings.HasPrefix(name, "__") {
		return true
	}

	return isBuiltinScalar(name)
}

func formatField(f *ast.FieldDefinition) string {
	var sb strings.Builder
	if f.Description != "" {
		sb.WriteString("\"\"\"" + f.Description + "\"\"\"\n")
	}

	sb.WriteString(f.Name)

	if len(f.Arguments) > 0 {
		sb.WriteString("(")

		for i, arg := range f.Arguments {
			if i > 0 {
				sb.WriteString(", ")
			}

			sb.WriteString(arg.Name)
			sb.WriteString(": ")
			sb.WriteString(arg.Type.String())

			if arg.DefaultValue != nil {
				sb.WriteString(" = ")
				sb.WriteString(arg.DefaultValue.String())
			}
		}

		sb.WriteString(")")
	}

	sb.WriteString(": ")
	sb.WriteString(f.Type.String())

	return sb.String()
}

func formatTypeDef(t *ast.Definition) string {
	var sb strings.Builder
	if t.Description != "" {
		sb.WriteString("\"\"\"" + t.Description + "\"\"\"\n")
	}

	switch t.Kind {
	case ast.Object:
		sb.WriteString("type " + t.Name)

		if len(t.Interfaces) > 0 {
			sb.WriteString(" implements ")
			sb.WriteString(strings.Join(t.Interfaces, " & "))
		}

		sb.WriteString(" {\n")

		for _, f := range t.Fields {
			if isBuiltinField(f.Name) {
				continue
			}

			if f.Description != "" {
				sb.WriteString("  \"\"\"" + f.Description + "\"\"\"\n")
			}

			sb.WriteString("  ")
			sb.WriteString(formatField(f))
			sb.WriteString("\n")
		}

		sb.WriteString("}")
	case ast.InputObject:
		sb.WriteString("input " + t.Name + " {\n")

		for _, f := range t.Fields {
			if f.Description != "" {
				sb.WriteString("  \"\"\"" + f.Description + "\"\"\"\n")
			}

			sb.WriteString("  " + f.Name + ": " + f.Type.String())

			if f.DefaultValue != nil {
				sb.WriteString(" = " + f.DefaultValue.String())
			}

			sb.WriteString("\n")
		}

		sb.WriteString("}")
	case ast.Enum:
		sb.WriteString("enum " + t.Name + " {\n")

		for _, v := range t.EnumValues {
			if v.Description != "" {
				sb.WriteString("  \"\"\"" + v.Description + "\"\"\"\n")
			}

			sb.WriteString("  " + v.Name + "\n")
		}

		sb.WriteString("}")
	case ast.Interface:
		sb.WriteString("interface " + t.Name + " {\n")

		for _, f := range t.Fields {
			if isBuiltinField(f.Name) {
				continue
			}

			if f.Description != "" {
				sb.WriteString("  \"\"\"" + f.Description + "\"\"\"\n")
			}

			sb.WriteString("  ")
			sb.WriteString(formatField(f))
			sb.WriteString("\n")
		}

		sb.WriteString("}")
	case ast.Union:
		sb.WriteString("union " + t.Name + " = ")
		sb.WriteString(strings.Join(t.Types, " | "))
	case ast.Scalar:
		sb.WriteString("scalar " + t.Name)
	}

	return sb.String()
}
