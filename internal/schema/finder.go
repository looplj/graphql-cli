package schema

import (
	"github.com/looplj/graphql-cli/pkg/graphql"
)

// FindScope is an alias for the public API type.
type FindScope = graphql.FindScope

// FindResult is an alias for the public API type.
type FindResult = graphql.FindResult

// ParseAndFind loads and searches the schema. It delegates to the public API.
func ParseAndFind(sdl string, keyword string, scope FindScope) ([]FindResult, error) {
	c := graphql.New()

	s, err := c.LoadSchemaFromSDL("parse-and-find", sdl)
	if err != nil {
		return nil, err
	}

	return graphql.Find(s, keyword, scope)
}
