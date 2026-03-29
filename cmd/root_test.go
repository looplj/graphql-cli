package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestEndpointFlagScope(t *testing.T) {
	if rootCmd.PersistentFlags().Lookup("endpoint") != nil {
		t.Fatal("expected endpoint flag to be scoped to execution commands, not root")
	}

	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "query", cmd: queryCmd},
		{name: "mutate", cmd: mutateCmd},
		{name: "find", cmd: findCmd},
	} {
		flag := tc.cmd.Flags().Lookup("endpoint")
		if flag == nil {
			t.Fatalf("expected %s command to define --endpoint", tc.name)
		}

		if flag.Shorthand != "e" {
			t.Fatalf("expected %s command to keep -e shorthand, got %q", tc.name, flag.Shorthand)
		}
	}
}
