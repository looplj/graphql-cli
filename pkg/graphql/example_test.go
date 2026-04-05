package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/looplj/graphql-cli/pkg/graphql"
)

func Example() {
	client := graphql.New()

	// Load a schema from SDL string.
	schema, err := client.LoadSchemaFromSDL("example", `
		type Query {
			users(limit: Int): [User!]!
			user(id: ID!): User
		}

		type User {
			id: ID!
			name: String!
			email: String!
			role: Role!
		}

		enum Role {
			ADMIN
			MEMBER
			GUEST
		}

		input CreateUserInput {
			name: String!
			email: String!
			role: Role = MEMBER
		}

		type Mutation {
			createUser(input: CreateUserInput!): User!
		}
	`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Schema loaded, SDL length:", len(schema.SDL))

	// Search for all definitions matching "user".
	results, err := graphql.Find(schema, "user", graphql.FindScope{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound %d results for \"user\":\n", len(results))

	for _, r := range results {
		fmt.Printf("  [%s] %s\n", r.Kind, r.Name)
	}

	// Search only queries.
	queries, _ := graphql.Find(schema, "", graphql.FindScope{Query: true})
	fmt.Printf("\nAll queries (%d):\n", len(queries))

	for _, q := range queries {
		fmt.Printf("  %s\n", q.Definition)
	}

	// Search only enums.
	enums, _ := graphql.Find(schema, "", graphql.FindScope{Enum: true})
	fmt.Printf("\nAll enums (%d):\n", len(enums))

	for _, e := range enums {
		fmt.Printf("  %s\n", e.Name)
	}
}

// This example shows how to execute a GraphQL query against a remote endpoint.
func Example_execute() {
	client := graphql.New()

	resp, err := client.Execute(context.Background(), "https://api.example.com/graphql",
		graphql.Request{
			Query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			Variables: map[string]any{
				"id": "123",
			},
		},
		map[string]string{
			"Authorization": "Bearer my-token",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	if resp.HasErrors() {
		log.Fatalf("GraphQL errors: %s", resp.Error())
	}

	var data struct {
		User struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: %s (%s)\n", data.User.Name, data.User.ID)
}
