package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/looplj/graphql-cli/internal/auth"
	"github.com/looplj/graphql-cli/internal/config"
)

const introspectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args { ...InputValue }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args { ...InputValue }
    type { ...TypeRef }
    isDeprecated
    deprecationReason
  }
  inputFields { ...InputValue }
  interfaces { ...TypeRef }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes { ...TypeRef }
}

fragment InputValue on __InputValue {
  name
  description
  type { ...TypeRef }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
            }
          }
        }
      }
    }
  }
}
`

func LoadSDL(ep *config.Endpoint) (string, error) {
	if ep.SchemaFile != "" {
		return loadFromFile(ep.SchemaFile)
	}

	if ep.URL != "" {
		return loadFromIntrospection(ep)
	}

	return "", fmt.Errorf("endpoint %q has neither url nor schema_file", ep.Name)
}

func loadFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read schema file: %w", err)
	}

	return string(data), nil
}

func loadFromIntrospection(ep *config.Endpoint) (string, error) {
	body, err := json.Marshal(map[string]string{"query": introspectionQuery})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ep.URL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range ep.Headers {
		req.Header.Set(k, v)
	}

	// Apply stored credentials (from `graphql-cli login`)
	store := auth.NewStore()
	if cred, _ := store.Load(ep.Name); cred != nil {
		for k, v := range cred.AuthHeaders() {
			req.Header.Set(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("introspection request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("introspection failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Schema json.RawMessage `json:"__schema"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse introspection response: %w", err)
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("introspection error: %s", result.Errors[0].Message)
	}

	return introspectionToSDL(result.Data.Schema)
}

func introspectionToSDL(schemaJSON json.RawMessage) (string, error) {
	var schema introspectionSchema
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return "", fmt.Errorf("parse schema: %w", err)
	}

	var buf bytes.Buffer

	for _, t := range schema.Types {
		if len(t.Name) > 0 && t.Name[0] == '_' && len(t.Name) > 1 && t.Name[1] == '_' {
			continue
		}

		switch t.Kind {
		case "SCALAR":
			if isBuiltinScalar(t.Name) {
				continue
			}

			writeDescription(&buf, t.Description, "")
			fmt.Fprintf(&buf, "scalar %s\n\n", t.Name)
		case "OBJECT":
			writeDescription(&buf, t.Description, "")
			fmt.Fprintf(&buf, "type %s", t.Name)

			if len(t.Interfaces) > 0 {
				buf.WriteString(" implements ")

				for i, iface := range t.Interfaces {
					if i > 0 {
						buf.WriteString(" & ")
					}

					buf.WriteString(resolveTypeName(iface))
				}
			}

			buf.WriteString(" {\n")

			for _, f := range t.Fields {
				writeDescription(&buf, f.Description, "  ")
				fmt.Fprintf(&buf, "  %s", f.Name)
				writeArgs(&buf, f.Args)
				fmt.Fprintf(&buf, ": %s", resolveTypeRef(f.Type))

				if f.IsDeprecated {
					fmt.Fprintf(&buf, " @deprecated")

					if f.DeprecationReason != "" {
						fmt.Fprintf(&buf, "(reason: %q)", f.DeprecationReason)
					}
				}

				buf.WriteString("\n")
			}

			buf.WriteString("}\n\n")
		case "INPUT_OBJECT":
			writeDescription(&buf, t.Description, "")
			fmt.Fprintf(&buf, "input %s {\n", t.Name)

			for _, f := range t.InputFields {
				writeDescription(&buf, f.Description, "  ")
				fmt.Fprintf(&buf, "  %s: %s", f.Name, resolveTypeRef(f.Type))

				if f.DefaultValue != "" {
					fmt.Fprintf(&buf, " = %s", f.DefaultValue)
				}

				buf.WriteString("\n")
			}

			buf.WriteString("}\n\n")
		case "ENUM":
			writeDescription(&buf, t.Description, "")
			fmt.Fprintf(&buf, "enum %s {\n", t.Name)

			for _, v := range t.EnumValues {
				writeDescription(&buf, v.Description, "  ")
				fmt.Fprintf(&buf, "  %s", v.Name)

				if v.IsDeprecated {
					fmt.Fprintf(&buf, " @deprecated")

					if v.DeprecationReason != "" {
						fmt.Fprintf(&buf, "(reason: %q)", v.DeprecationReason)
					}
				}

				buf.WriteString("\n")
			}

			buf.WriteString("}\n\n")
		case "INTERFACE":
			writeDescription(&buf, t.Description, "")
			fmt.Fprintf(&buf, "interface %s {\n", t.Name)

			for _, f := range t.Fields {
				writeDescription(&buf, f.Description, "  ")
				fmt.Fprintf(&buf, "  %s", f.Name)
				writeArgs(&buf, f.Args)
				fmt.Fprintf(&buf, ": %s\n", resolveTypeRef(f.Type))
			}

			buf.WriteString("}\n\n")
		case "UNION":
			writeDescription(&buf, t.Description, "")
			fmt.Fprintf(&buf, "union %s = ", t.Name)

			for i, pt := range t.PossibleTypes {
				if i > 0 {
					buf.WriteString(" | ")
				}

				buf.WriteString(resolveTypeName(pt))
			}

			buf.WriteString("\n\n")
		}
	}

	return buf.String(), nil
}

type introspectionSchema struct {
	QueryType        *introspectionNameRef    `json:"queryType"`
	MutationType     *introspectionNameRef    `json:"mutationType"`
	SubscriptionType *introspectionNameRef    `json:"subscriptionType"`
	Types            []introspectionType      `json:"types"`
	Directives       []introspectionDirective `json:"directives"`
}

type introspectionNameRef struct {
	Name string `json:"name"`
}

type introspectionType struct {
	Kind          string                 `json:"kind"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Fields        []introspectionField   `json:"fields"`
	InputFields   []introspectionInput   `json:"inputFields"`
	Interfaces    []introspectionTypeRef `json:"interfaces"`
	EnumValues    []introspectionEnum    `json:"enumValues"`
	PossibleTypes []introspectionTypeRef `json:"possibleTypes"`
}

type introspectionField struct {
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	Args              []introspectionInput `json:"args"`
	Type              introspectionTypeRef `json:"type"`
	IsDeprecated      bool                 `json:"isDeprecated"`
	DeprecationReason string               `json:"deprecationReason"`
}

type introspectionInput struct {
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Type         introspectionTypeRef `json:"type"`
	DefaultValue string               `json:"defaultValue"`
}

type introspectionEnum struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
}

type introspectionTypeRef struct {
	Kind   string                `json:"kind"`
	Name   string                `json:"name"`
	OfType *introspectionTypeRef `json:"ofType"`
}

type introspectionDirective struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Locations   []string             `json:"locations"`
	Args        []introspectionInput `json:"args"`
}

func resolveTypeRef(t introspectionTypeRef) string {
	switch t.Kind {
	case "NON_NULL":
		if t.OfType != nil {
			return resolveTypeRef(*t.OfType) + "!"
		}
	case "LIST":
		if t.OfType != nil {
			return "[" + resolveTypeRef(*t.OfType) + "]"
		}
	default:
		return t.Name
	}

	return "Unknown"
}

func resolveTypeName(t introspectionTypeRef) string {
	if t.Name != "" {
		return t.Name
	}

	if t.OfType != nil {
		return resolveTypeName(*t.OfType)
	}

	return "Unknown"
}

func writeDescription(buf *bytes.Buffer, desc, indent string) {
	if desc == "" {
		return
	}

	fmt.Fprintf(buf, "%s\"\"\"%s\"\"\"\n", indent, desc)
}

func writeArgs(buf *bytes.Buffer, args []introspectionInput) {
	if len(args) == 0 {
		return
	}

	buf.WriteString("(")

	for i, a := range args {
		if i > 0 {
			buf.WriteString(", ")
		}

		fmt.Fprintf(buf, "%s: %s", a.Name, resolveTypeRef(a.Type))

		if a.DefaultValue != "" {
			fmt.Fprintf(buf, " = %s", a.DefaultValue)
		}
	}

	buf.WriteString(")")
}

func isBuiltinScalar(name string) bool {
	switch name {
	case "String", "Int", "Float", "Boolean", "ID":
		return true
	}

	return false
}
