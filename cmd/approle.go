package cmd

import (
	"context"

	"github.com/guardanet/kubectl-vaultlogin/pkg/federate"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// variable to store provided mount point path for Vault's approle authentication
var VaultApproleAuthMount string

// vaultApproleMountPath defines Vault's approle mount point
const defaultVaultApproleMountPoint = "/approle"

// Approle() creates an approle cobra subcommand
// test bool is used to designate if the instance is a test run (true) or actual request (false)
func Approle(test bool) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "approle",
		Args:  cobra.NoArgs,
		Short: "Authenticates to Hashicorp Vault using approle authentication",
		Long: `Authenticates to Hashicorp Vault using approle authentication.
It expects the Role ID and Secret ID to be suupplied in environemtn variables:
APPROLE_ROLE_ID and APPROLE_SECRET_ID respectively.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// In order to differentiate between "cobra command line" errors and actual program errors
			// as well as print Usage ONLY when the error results from imnproper command specification
			// and not from the actual program, we do the following:
			// 		1. check if all required flags are set (here 2 out of 3 flags get a default value so we only check 1 flag)
			//			a. if all is good then we switch OFF SilenceUsage and SilenceErrors and proceed with command execution
			// 			b. otherwise we return with the actual error cmd.Context().Err()
			// SilenceUsage:  true,
			// SilenceErrors: true,

			psatArgs := viper.GetViper().AllSettings()
			if psatArgs[federate.FlagVaultApproleAuthMount].(string) != "" &&
				psatArgs["vault-address"].(string) != "" {

				cmd.SilenceUsage = true
				cmd.SilenceErrors = true

				ctx := context.Background()
				return federate.FederateWithApprole(&ctx, psatArgs, test)

			} else {
				return cmd.Context().Err()
			}
		},
	}

	// init - configure flags that only apply to this subcommand and its children (if any)
	cmd.PersistentFlags().StringVarP(&VaultApproleAuthMount, federate.FlagVaultApproleAuthMount, "a", defaultVaultApproleMountPoint, "vault approle authentication mountpoint, ex: /approle")
	// cmd.MarkPersistentFlagRequired(federate.FlagVaultApproleAuthMount)
	viper.BindPFlag(federate.FlagVaultApproleAuthMount, cmd.PersistentFlags().Lookup(federate.FlagVaultApproleAuthMount))

	return cmd
}
