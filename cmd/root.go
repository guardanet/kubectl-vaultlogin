package cmd

import (
	"errors"
	"log"
	"os"

	kvlerrors "github.com/guardanet/kubectl-vaultlogin/pkg/errors"
	"github.com/guardanet/kubectl-vaultlogin/pkg/federate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/release-utils/version"
)

var VaultAddress string
var DownstreamClusterName string

// New() creates a new cobra Root Command
// test bool is used to designate if the instance is a test run (true) or actual request (false)
func New(test bool) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "kubectl-vaultlogin",
		Aliases: []string{"kubectl vaultlogin"},
		Args:    cobra.NoArgs,
		Short:   "client-go credential plugin for kubectl",
		Long: `kubectl-vaultlogin is a client-go credential plugin for kubectl,
that leverages Hashicorp Vault to request just-in-time short-lived kubernetes bearer tokens.
It expects an ExecCredential to be passed in KUBERNETES_EXEC_INFO environment variable
and produces a corresponding ExecCredential with a token and expiration, which it prints to STDOUT`,
	}

	// init
	cmd.PersistentFlags().StringVarP(&VaultAddress, federate.FlagVaultAddress, "v", "https://localhost:8200", "full URL with port to Hashicorp Vault.")
	cmd.MarkFlagRequired(federate.FlagVaultAddress)
	viper.BindPFlag(federate.FlagVaultAddress, cmd.PersistentFlags().Lookup(federate.FlagVaultAddress))

	cmd.PersistentFlags().StringVarP(&DownstreamClusterName, federate.FlagClusterName, "c", "", "a downstream cluster name, this must be consistent with the name that is used in the kubernetes secret engine path /kubernetes/<clustername>")
	viper.BindPFlag(federate.FlagClusterName, cmd.PersistentFlags().Lookup(federate.FlagClusterName))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	// Add subcommands
	cmd.AddCommand(Federate(test))
	cmd.AddCommand(version.WithFont(""))

	return cmd
}

func Execute() {
	log.SetPrefix("Error [kubectl-vaultlogin]: ")
	// log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lmsgprefix)
	log.SetFlags(log.Lmsgprefix)
	var kvlErr kvlerrors.KvlError
	// launching New command and passing false to indicate this is not a test run but a legitimate request
	if err := New(false).Execute(); err != nil {
		if errors.As(err, &kvlErr) {
			log.Fatalf("%v", err)
		} else {
			os.Exit(2)
		}
	}
}
