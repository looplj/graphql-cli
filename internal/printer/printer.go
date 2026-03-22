package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/looplj/graphql-cli/internal/schema"
)

var (
	kindColor = color.New(color.FgCyan, color.Bold)
	nameColor = color.New(color.FgGreen, color.Bold)
	dimColor  = color.New(color.FgHiBlack)
)

func PrintFindResults(results []schema.FindResult) {
	if len(results) == 0 {
		dimColor.Println("No results found.")
		return
	}

	for i, r := range results {
		if i > 0 {
			fmt.Println()
			dimColor.Println(strings.Repeat("─", 60))
			fmt.Println()
		}

		kindColor.Printf("[%s] ", strings.ToUpper(r.Kind))
		nameColor.Println(r.Name)
		fmt.Println()
		fmt.Println(r.Definition)
	}
}

func PrintJSON(data json.RawMessage) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}

	fmt.Println(pretty.String())
}

type GraphQLError struct {
	Message string `json:"message"`
}

func PrintErrors(errors []GraphQLError) {
	errColor := color.New(color.FgRed, color.Bold)
	for _, e := range errors {
		errColor.Printf("Error: ")
		fmt.Println(e.Message)
	}
}
