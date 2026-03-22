---
name: graphql-cli
description: "Manages GraphQL endpoints and executes queries/mutations using the graphql-cli tool. Use when asked to query a GraphQL API, explore a GraphQL schema, add/manage endpoints, or authenticate with a GraphQL service."
license: MIT
compatibility: Requires graphql-cli binary (Go) built and available in PATH
metadata:
  author: looplj
  version: "1.0"
---

# GraphQL CLI

A skill for managing GraphQL endpoints and executing operations using `graphql-cli`.

## Capabilities

- Add and manage multiple GraphQL endpoints (remote URL or local schema file)
- Authenticate with endpoints (Bearer token, Basic auth, custom header)
- Execute GraphQL queries and mutations
- Explore and search GraphQL schemas by keyword
- List configured endpoints

## Prerequisites

Run with npx (no installation required):

```bash
npx @axonhub/graphql-cli <command>
```

Or build from source:

```bash
go install github.com/looplj/graphql-cli@latest
```

## Workflows

### 1. Add an endpoint

**Remote URL:**
```bash
graphql-cli add <name> --url <graphql-url> [--description "desc"] [--header "Key=Value"]
```

**Local schema file:**
```bash
graphql-cli add <name> --schema-file ./schema.graphql [--description "desc"]
```

Example:
```bash
graphql-cli add production --url https://api.example.com/graphql --description "Prod API"
graphql-cli add local --schema-file ./testdata/schema.graphql --description "Local schema"
```

### 2. List endpoints

```bash
graphql-cli list            # names and URLs
graphql-cli list --detail   # includes headers (masked) and auth status
```

### 3. Authenticate

```bash
# Interactive (prompts for auth type and credentials)
graphql-cli login <endpoint>

# Non-interactive
graphql-cli login <endpoint> --type token --token "my-api-key"
graphql-cli login <endpoint> --type basic --user admin --pass secret
graphql-cli login <endpoint> --type header --key X-API-Key --value "key123"

# Remove credentials
graphql-cli logout <endpoint>
```

Credentials are stored in the OS keyring (macOS Keychain, Windows Credential Manager, GNOME Keyring) with a plaintext file fallback.

You can also specify the endpoint via `-e`:
```bash
graphql-cli login -e production --type token --token "my-token"
```

### 4. Execute a query

```bash
graphql-cli query '<graphql-query>' -e <endpoint>
graphql-cli query -f query.graphql -e <endpoint>
graphql-cli query '{ user(id: "1") { name } }' -e <endpoint> -v '{"id": "1"}'
graphql-cli query '{ me { name } }' -e <endpoint> -H "Authorization=Bearer token"
```

### 5. Execute a mutation

```bash
graphql-cli mutate '<graphql-mutation>' -e <endpoint>
graphql-cli mutate -f mutation.graphql -e <endpoint> -v '{"name": "test"}'
graphql-cli mutate 'mutation { createUser(name: "test") { id } }' -e <endpoint>
```

### 6. Explore the schema

```bash
# Search all definitions
graphql-cli find <keyword> -e <endpoint>

# Narrow by kind
graphql-cli find user -e <endpoint> --query          # Query fields only
graphql-cli find user -e <endpoint> --mutation        # Mutation fields only
graphql-cli find user -e <endpoint> --type            # Object/Interface/Union/Scalar types
graphql-cli find user -e <endpoint> --input           # Input types only
graphql-cli find status -e <endpoint> --enum          # Enum types only

# List everything (no keyword)
graphql-cli find -e <endpoint>

# Combine scopes
graphql-cli find user -e <endpoint> --type --input
```

Schema is loaded via introspection (remote URL) or from a local file (schema_file).

## Header priority

When executing queries/mutations, headers are merged with this priority (highest wins):

1. CLI `-H` flags
2. Stored credentials (`login`)
3. Config file headers

## Common patterns

### Query with variables from a file
```bash
graphql-cli query -f queries/get-user.graphql -e prod -v "$(cat vars.json)"
```

### Pipe output to jq
```bash
graphql-cli query '{ users { id name } }' -e prod 2>/dev/null | jq '.users[0]'
```

### Explore before querying
```bash
# First, find what queries are available
graphql-cli find -e prod --query

# Then find the input types needed
graphql-cli find CreateUser -e prod --input

# Then execute
graphql-cli mutate 'mutation { createUser(input: {name: "Alice", email: "alice@example.com"}) { id } }' -e prod
```