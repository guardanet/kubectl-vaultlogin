package cmd

import (
	"context"

	"github.com/guardanet/kubectl-vaultlogin/pkg/federate"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var VaultKubernetesAuthMount string
var PsatPath string

// Psat() creates a psat cobra subcommand
// test bool is used to designate if the instance is a test run (true) or actual request (false)
func Psat(test bool) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "psat",
		Args:  cobra.NoArgs,
		Short: "Authenticates to Hashicorp Vault using kubernetes authentication",
		Long: `Authenticates to Hashicorp Vault using kubernetes authentication
	and its projected service account token.`,
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
			if psatArgs[federate.FlagVaultKubernetesAuthMount].(string) != "" &&
				psatArgs[federate.FlagPsatPath].(string) != "" &&
				// global
				psatArgs["vault-address"].(string) != "" {

				cmd.SilenceUsage = true
				cmd.SilenceErrors = true
				ctx := context.Background()
				// FederateWithPsat performs actual end to end federation using PSAT
				// test is a boolean and if true means this is a test run and not actual request
				return federate.FederateWithPsat(&ctx, psatArgs, test)
			}
			return cmd.Context().Err()
		},
	}

	// init - configure flags that only apply to this subcommand and its children (if any)
	cmd.PersistentFlags().StringVarP(&VaultKubernetesAuthMount, federate.FlagVaultKubernetesAuthMount, "a", "", "vault kuberentes authentication mountpoint, ex: /kubernetes/<clustername>")
	cmd.MarkPersistentFlagRequired(federate.FlagVaultKubernetesAuthMount)
	viper.BindPFlag(federate.FlagVaultKubernetesAuthMount, cmd.PersistentFlags().Lookup(federate.FlagVaultKubernetesAuthMount))

	cmd.PersistentFlags().StringVarP(&PsatPath, federate.FlagPsatPath, "p", "/var/run/secrets/kubernetes.io/serviceaccount/token", "absolute path to projected service account token used to authenticate to hashicorp vault")
	// cmd.MarkPersistentFlagRequired(federate.FlagPsatPath)
	viper.BindPFlag(federate.FlagPsatPath, cmd.PersistentFlags().Lookup(federate.FlagPsatPath))

	return cmd
}
