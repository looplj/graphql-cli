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
graphql-cli endpoint add production --url https://api.example.com/graphql

# Add a local schema endpoint
graphql-cli endpoint add local --schema-file ./schema.graphql

# Execute a query
graphql-cli query -e production '{ users { id name } }'

# Explore the schema
graphql-cli find -e production user
```

## Commands

### `endpoint` — Manage endpoints

```bash
graphql-cli endpoint <subcommand>
```

Available subcommands:

- `add` — add a new endpoint
- `list` — list configured endpoints
- `update` — update an existing endpoint
- `login` — store credentials for an endpoint
- `logout` — remove stored credentials for an endpoint

### `endpoint add` — Add a new endpoint

```bash
graphql-cli endpoint add <name> --url <url> [--schema-file <path>] [-d <description>] [--header key=value]
```

**Examples:**

```bash
graphql-cli endpoint add production --url https://api.example.com/graphql --header "Authorization=Bearer token"
graphql-cli endpoint add local --schema-file ./schema.graphql --description "Local dev schema"
```

### `endpoint list` — List configured endpoints

```bash
graphql-cli endpoint list [--detail]
```

Use `--detail` to show headers, schema file paths, and auth status.

Example:

```bash
graphql-cli endpoint list --detail
```

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

### `endpoint update` — Update an existing endpoint

```bash
graphql-cli endpoint update <name> [--url <url>] [-d <description>] [--header key=value]
```

**Examples:**

```bash
graphql-cli endpoint update production --url https://api.example.com/v2/graphql
graphql-cli endpoint update production --header "Authorization=Bearer new-token"
graphql-cli endpoint update production --url https://new-url.com/graphql --header "X-Custom=value" -d "Updated endpoint"
```

### `find` — Search schema definitions

```bash
graphql-cli find -e <endpoint> [keyword] [--query] [--mutation] [--type] [--input] [--enum] [--detail]
```

By default, only names are shown. Use `--detail` to display full definitions with fields and arguments.

The keyword supports glob syntax (`*`, `?`, `[...]`). Without glob characters, it matches as a substring (e.g., `user` matches `getUser`, `UserInput`).

**Examples:**

```bash
graphql-cli find -e production user
graphql-cli find -e production "get*"
graphql-cli find -e production user --query
graphql-cli find -e production --mutation
graphql-cli find -e production status --enum
graphql-cli find -e production user --detail
```

### `endpoint login` — Authenticate with an endpoint

Credentials are stored in the OS keyring (macOS Keychain, Windows Credential Manager, GNOME Keyring) with a plaintext file fallback.

```bash
graphql-cli endpoint login <endpoint> [--type token|basic|header]
```

**Supported auth types:**

| Type     | Description                          |
|----------|--------------------------------------|
| `token`  | Bearer token (API key, JWT, PAT)     |
| `basic`  | Username + password (Basic Auth)     |
| `header` | Custom header key=value              |

**Examples:**

```bash
graphql-cli endpoint login production
graphql-cli endpoint login production --type token --token "my-token"
```

### `endpoint logout` — Remove stored credentials

```bash
graphql-cli endpoint logout <endpoint>
```

### `audit list` — List recorded queries and mutations

```bash
graphql-cli audit list [--endpoint <name>] [--status success|error] [--contains <text>] [--query|--mutation] [--detail] [--limit <n>]
```

Examples:

```bash
graphql-cli audit list
graphql-cli audit list --endpoint production
graphql-cli audit list --status error
graphql-cli audit list --contains createUser
graphql-cli audit list --mutation --detail
```

## Configuration

The configuration file is stored at `~/.config/graphql-cli/config.yaml` by default. Use `--config` to specify a custom path.

## Audit Log

Executed GraphQL statements are appended to `~/.config/graphql-cli/audit.log` as JSON lines.

Each line records:

- `timestamp`
- `endpoint`
- `url`
- `status`
- `statement`
- `error` when execution fails

Example entry:

```json
{"timestamp":"2026-03-29T08:15:30.123456Z","endpoint":"production","url":"https://api.example.com/graphql","status":"success","statement":"query { viewer { id } }"}
```

You can inspect recorded operations with:

```bash
graphql-cli audit list
graphql-cli audit list --query
graphql-cli audit list --status error
graphql-cli audit list --contains viewer
graphql-cli audit list --mutation --detail
```

Or stream the raw log with:

```bash
tail -f ~/.config/graphql-cli/audit.log
```

## Global Flags

| Flag                  | Description                          |
|-----------------------|--------------------------------------|
| `--config <path>`     | Config file path                     |

## License

[MIT](LICENSE)
