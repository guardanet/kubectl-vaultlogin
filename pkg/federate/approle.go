package federate

import (
	"context"
	"fmt"
	"os"

	kvlerrors "github.com/guardanet/kubectl-vaultlogin/pkg/errors"
)

// const to define cobra command flag name that supplies mount point path for Vault's approle authentication
const FlagVaultApproleAuthMount = "vault-approle-auth-mount"

// FederateWithApprole() perfoms all actions resulting from the federate approle subcommand to request a new kubernetes bearer token
// and responds with a corresponding ExecCredetnial written to STDOUT
// ctx - is context created by cobra subcommand RunE function
// args - are args passed from the cobra subcommand RunE function
// test - indicates if this is a test run (true) or actual request (false). If true no communication with Vault is perfomed, instead a fake beaer token is created
func FederateWithApprole(ctx *context.Context, args map[string]any, test bool) error {
	// fmt.Printf("Path to PSAT: %s\nAuth Mount Point: %s\n", args["psat-path"], args[FlagVaultApproleAuthMount])

	// verify supplied flags
	// check vault-address
	if err := isValidURL(args["vault-address"].(string)); err != nil {
		return kvlerrors.New(err.Error())
	}
	// check vault mount points
	if !isAbsolutePath(args[FlagVaultApproleAuthMount].(string)) {
		return kvlerrors.New(fmt.Errorf("vault-approle-auth-mount must be of a form of an absolute path with alphanumeric path elements, ex. /approle: %s", args[FlagVaultApproleAuthMount].(string)).Error())
	}

	// perform preparation tasks
	if err := prepFederation(&args); err != nil {
		return kvlerrors.New(err.Error())
	}

	// only actual resques and not a test run
	if !test {
		var err error
		// now authenticate to vault leveraging PSAT
		if err = prepVaultClient(args["vault-address"].(string)); err != nil {
			return kvlerrors.New(err.Error())
		}
		if err = authToVaultWithApprole(ctx, client, args[FlagVaultApproleAuthMount].(string)); err != nil {
			return kvlerrors.New(err.Error())
		}

		// now that we are authenticated, lets generate a token to authenticate to k8s cluster
		k8sToken, err = generateK8sToken(ctx, client, vaultK8sSecretRole, args["cluster-name"].(string))
		if err != nil {
			return kvlerrors.New(err.Error())
		}
	} else {
		k8sToken = generateFakeK8sToken()
	}

	// Assemble a ExecCredential
	err := assembleExecCredential(inputExecCredentialPtr, k8sToken)
	if err != nil {
		return kvlerrors.New(err.Error())
	}

	// Output the new ExecCredential
	err = printExecCredential(os.Stdout, inputExecCredentialPtr)
	if err != nil {
		return kvlerrors.New(err.Error())
	}
	return nil
}
