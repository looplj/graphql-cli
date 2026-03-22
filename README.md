# graphql-cli

A CLI tool for exploring and querying GraphQL APIs. Supports configuring multiple endpoints (remote URL or local schema file), schema exploration, and executing queries/mutations.

## Installation

### From source

```bash
go install github.com/looplj/graphql-cli@latest
```

### From releases

Download the binary for your platform from the [Releases](https://github.com/looplj/graphql-cli/releases) page.

## Quick Start

```bash
# Add a remote endpoint
graphql-cli add production --url https://api.example.com/graphql

# Add a local schema endpoint
graphql-cli add local --schema-file ./schema.graphql

# Execute a query
graphql-cli query -e production '{ users { id name } }'

# Explore the schema
graphql-cli find -e production user
```

## Commands

### `add` — Add a new endpoint

```bash
graphql-cli add <name> --url <url> [--schema-file <path>] [-d <description>] [--header key=value]
```

**Examples:**

```bash
graphql-cli add production --url https://api.example.com/graphql --header "Authorization=Bearer token"
graphql-cli add local --schema-file ./schema.graphql --description "Local dev schema"
```

### `list` — List configured endpoints

```bash
graphql-cli list [--detail]
```

Use `--detail` to show headers, schema file paths, and auth status.

### `query` — Execute a GraphQL query

```bash
graphql-cli query -e <endpoint> '<query>' [-f <file>] [-v '<variables>'] [-H key=value]
```

**Examples:**

```bash
graphql-cli query -e production '{ users { id name } }'
graphql-cli query -e production -f query.graphql
graphql-cli query -e production '{ user(id: "1") { name } }' -v '{"id": "1"}'
graphql-cli query -e production '{ me { name } }' -H "Authorization=Bearer token"
```

### `mutate` — Execute a GraphQL mutation

```bash
graphql-cli mutate -e <endpoint> '<mutation>' [-f <file>] [-v '<variables>'] [-H key=value]
```

**Examples:**

```bash
graphql-cli mutate -e production 'mutation { createUser(name: "test") { id } }'
graphql-cli mutate -e production -f mutation.graphql -v '{"name": "test"}'
```

### `find` — Search schema definitions

```bash
graphql-cli find -e <endpoint> [keyword] [--query] [--mutation] [--type] [--input] [--enum]
```

**Examples:**

```bash
graphql-cli find -e production user
graphql-cli find -e production user --query
graphql-cli find -e production --mutation
graphql-cli find -e production status --enum
```

### `login` — Authenticate with an endpoint

Credentials are stored in the OS keyring (macOS Keychain, Windows Credential Manager, GNOME Keyring) with a plaintext file fallback.

```bash
graphql-cli login [endpoint] [-e <endpoint>] [--type token|basic|header]
```

**Supported auth types:**

| Type     | Description                          |
|----------|--------------------------------------|
| `token`  | Bearer token (API key, JWT, PAT)     |
| `basic`  | Username + password (Basic Auth)     |
| `header` | Custom header key=value              |

**Examples:**

```bash
graphql-cli login production
graphql-cli login production --type token --token "my-token"
```

### `logout` — Remove stored credentials

```bash
graphql-cli logout [endpoint] [-e <endpoint>]
```

## Configuration

The configuration file is stored at `~/.config/graphql-cli/config.yaml` by default. Use `--config` to specify a custom path.

## Global Flags

| Flag                  | Description                          |
|-----------------------|--------------------------------------|
| `--config <path>`     | Config file path                     |
| `-e, --endpoint <name>` | Endpoint name to use              |

## License

[MIT](LICENSE)
