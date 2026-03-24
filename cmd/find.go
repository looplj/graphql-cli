package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/looplj/graphql-cli/internal/config"
	"github.com/looplj/graphql-cli/internal/printer"
	"github.com/looplj/graphql-cli/internal/schema"
)

var findCmd = &cobra.Command{
	Use:   "find [keyword]",
	Short: "Search schema definitions by keyword",
	Long: `Search for types, queries, mutations, inputs, and enums in the GraphQL schema.
Use flags to narrow the search scope.

The keyword supports glob syntax (*, ?, [...]). Without glob characters,
it matches as a substring (e.g., "user" matches "getUser", "UserInput").

Examples:
  graphql-cli find user                              # substring match
  graphql-cli find "get*"                            # glob: starts with "get"
  graphql-cli find "User?"                           # glob: "User" + one char
  graphql-cli find "{createUser,CreateUserInput}"    # glob: exact alternatives
  graphql-cli find "[A-Z]*Input"                     # glob: capitalized, ends with "Input"
  graphql-cli find user --query                      # only Query fields
  graphql-cli find --mutation                        # list all mutations
  graphql-cli find user --type --input               # types and inputs
  graphql-cli find status --enum                     # enums only
  graphql-cli find user --detail                     # show full definitions`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: requireEndpoint,
	RunE:    runFind,
}

var (
	findQuery    bool
	findMutation bool
	findType     bool
	findInput    bool
	findEnum     bool
	findDetail   bool
)

func init() {
	findCmd.Flags().BoolVar(&findQuery, "query", false, "search only Query fields")
	findCmd.Flags().BoolVar(&findMutation, "mutation", false, "search only Mutation fields")
	findCmd.Flags().BoolVar(&findType, "type", false, "search only Object/Interface/Union/Scalar types")
	findCmd.Flags().BoolVar(&findInput, "input", false, "search only Input types")
	findCmd.Flags().BoolVar(&findEnum, "enum", false, "search only Enum types")
	findCmd.Flags().BoolVar(&findDetail, "detail", false, "show fields and arguments")
	rootCmd.AddCommand(findCmd)
}

func runFind(cmd *cobra.Command, args []string) error {
	keyword := ""
	if len(args) > 0 {
		keyword = strings.TrimSpace(args[0])
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	ep, err := cfg.GetEndpoint(endpointName)
	if err != nil {
		return err
	}

	sdl, err := schema.LoadSDL(ep)
	if err != nil {
		return err
	}

	scope := schema.FindScope{
		Query:    findQuery,
		Mutation: findMutation,
		Type:     findType,
		Input:    findInput,
		Enum:     findEnum,
	}

	results, err := schema.ParseAndFind(sdl, keyword, scope)
	if err != nil {
		return err
	}

	printer.PrintFindResults(results, findDetail)

	return nil
}
