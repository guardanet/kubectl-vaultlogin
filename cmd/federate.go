package cmd

import (
	"github.com/spf13/cobra"
)

// Federate() creates a federate cobra subcommand
// test bool is used to designate if the instance is a test run (true) or actual request (false)
func Federate(test bool) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "federate [command]",
		Args:  cobra.NoArgs,
		Short: "Federates an identity artifact with Hashicorp Vault to obtain a kubernetes beaer token",
		Long: `kubectl-vaultlogin federate federates an existing identity credential and exchanges it for a just-in-time short-lived kubernetes bearer token.
It supports two types of identity credentials, namely: 
- kubernetes projected service account tokens (PSATs) and 
- approle role-id/secret-id.`,
	}

	// init
	// nothing here

	// Add subcommands
	cmd.AddCommand(Psat(test))
	cmd.AddCommand(Approle(test))

	return cmd
}
