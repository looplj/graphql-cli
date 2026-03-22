package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/looplj/graphql-cli/internal/auth"
	"github.com/looplj/graphql-cli/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login [endpoint]",
	Short: "Authenticate with a GraphQL endpoint",
	Long: `Store credentials for a GraphQL endpoint. Credentials are saved
in the OS keyring (macOS Keychain, Windows Credential Manager, GNOME Keyring)
with a plaintext file fallback.

Supported auth types:
  token    - Bearer token (API key, JWT, PAT, etc.)
  basic    - Username + password (Basic Auth)
  header   - Custom header key=value

Examples:
  graphql-cli login                          # login to default endpoint
  graphql-cli login production               # login to specific endpoint
  graphql-cli login production --type token  # non-interactive with --token flag
  graphql-cli login production --type token --token "my-token"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogin,
}

var (
	loginType  string
	loginToken string
	loginUser  string
	loginPass  string
	loginKey   string
	loginValue string
)

func init() {
	loginCmd.Flags().StringVar(&loginType, "type", "", "auth type: token, basic, header")
	loginCmd.Flags().StringVar(&loginToken, "token", "", "bearer token (for --type token)")
	loginCmd.Flags().StringVar(&loginUser, "user", "", "username (for --type basic)")
	loginCmd.Flags().StringVar(&loginPass, "pass", "", "password (for --type basic)")
	loginCmd.Flags().StringVar(&loginKey, "key", "", "header key (for --type header)")
	loginCmd.Flags().StringVar(&loginValue, "value", "", "header value (for --type header)")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	epName, err := resolveEndpointName(args)
	if err != nil {
		return err
	}

	ep, err := cfg.GetEndpoint(epName)
	if err != nil {
		return err
	}

	cred, err := buildCredential()
	if err != nil {
		return err
	}

	store := auth.NewStore()

	insecure, err := store.Save(ep.Name, cred)
	if err != nil {
		return fmt.Errorf("save credential: %w", err)
	}

	successColor := color.New(color.FgGreen, color.Bold)
	successColor.Printf("✓ ")
	fmt.Printf("Logged in to %q as %s\n", ep.Name, cred.String())

	if insecure {
		warnColor := color.New(color.FgYellow)
		warnColor.Println("⚠ Stored in plaintext file (OS keyring unavailable)")
	}

	return nil
}

func buildCredential() (*auth.Credential, error) {
	authType := loginType
	if authType == "" {
		authType = promptSelect("Auth type", []string{"token", "basic", "header"})
	}

	switch authType {
	case "token":
		token := loginToken
		if token == "" {
			token = promptSecret("Token")
		}

		if token == "" {
			return nil, fmt.Errorf("token is required")
		}

		return &auth.Credential{Type: "token", Token: token}, nil

	case "basic":
		user := loginUser
		if user == "" {
			user = promptInput("Username")
		}

		pass := loginPass
		if pass == "" {
			pass = promptSecret("Password")
		}

		if user == "" || pass == "" {
			return nil, fmt.Errorf("username and password are required")
		}

		return &auth.Credential{Type: "basic", User: user, Pass: pass}, nil

	case "header":
		key := loginKey
		if key == "" {
			key = promptInput("Header key")
		}

		value := loginValue
		if value == "" {
			value = promptSecret("Header value")
		}

		if key == "" || value == "" {
			return nil, fmt.Errorf("header key and value are required")
		}

		return &auth.Credential{Type: "header", Key: key, Value: value}, nil

	default:
		return nil, fmt.Errorf("unknown auth type %q (supported: token, basic, header)", authType)
	}
}

func promptSelect(label string, options []string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s:\n", label)

	for i, opt := range options {
		fmt.Printf("  [%d] %s\n", i+1, opt)
	}

	fmt.Print("Choose (1-", len(options), "): ")

	input, _ := reader.ReadString('\n')

	input = strings.TrimSpace(input)
	for i, opt := range options {
		if input == fmt.Sprintf("%d", i+1) || input == opt {
			return opt
		}
	}

	if input == "" && len(options) > 0 {
		return options[0]
	}

	return input
}

func promptInput(label string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s: ", label)

	input, _ := reader.ReadString('\n')

	return strings.TrimSpace(input)
}

func promptSecret(label string) string {
	fmt.Printf("%s: ", label)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))

		fmt.Println()

		if err != nil {
			return ""
		}

		return strings.TrimSpace(string(b))
	}
	// non-terminal (piped input)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	return strings.TrimSpace(input)
}
