package federate

import (
	"context"
	"fmt"
	"os"

	kvlerrors "github.com/guardanet/kubectl-vaultlogin/pkg/errors"
)

// const to define cobra command flag name that supplies mount point path for Vault's kubernetes authentication
const FlagVaultKubernetesAuthMount = "vault-kubernetes-auth-mount"

// const to define cobra command flag name that supplies path to PSAT
const FlagPsatPath = "psat-path"

// FederateWithPsat() perfoms all actions resulting from the federate psat subcommand to request a new kubernetes bearer token
// and responds with a corresponding ExecCredetnial written to STDOUT
// ctx - is context created by cobra subcommand RunE function
// args - are args passed from the cobra subcommand RunE function
// test - indicates if this is a test run (true) or actual request (false). If true no communication with Vault is perfomed, instead a fake beaer token is created
func FederateWithPsat(ctx *context.Context, args map[string]any, test bool) error {
	// fmt.Printf("Path to PSAT: %s\nAuth Mount Point: %s\n", args[FlagPsatPath], args[FlagVaultKubernetesAuthMount])

	// verify supplied flags
	// check vault-address
	if err := isValidURL(args["vault-address"].(string)); err != nil {
		return kvlerrors.New(err.Error())
	}
	// check vault mount point
	if !isAbsolutePath(args[FlagVaultKubernetesAuthMount].(string)) {
		return kvlerrors.New(fmt.Errorf("vault-kubernetes-auth-mount must be of a form of an absolute path with alphanumeric path elements, ex. /kubernetes/argocd : %s", args[FlagVaultKubernetesAuthMount].(string)).Error())

	}
	if !isAbsolutePath(args[FlagPsatPath].(string)) {
		return kvlerrors.New(fmt.Errorf("psat-path must be an absolute path to a token file: %s", args[FlagPsatPath].(string)).Error())
	}

	// perform preparation tasks
	if err := prepFederation(&args); err != nil {
		return kvlerrors.New(err.Error())
	}

	// ensure kubernetes authentication mont point is properly set
	if err := prepVaultAuthMount(&args); err != nil {
		return kvlerrors.New(err.Error())
	}

	// only actual resques and not a test run
	if !test {
		var err error
		// now authenticate to vault leveraging PSAT
		if err := prepVaultClient(args["vault-address"].(string)); err != nil {
			return kvlerrors.New(err.Error())
		}
		if err := authToVaultWithKubernetes(ctx, client, vaultKubernetesLoginRole, args[FlagVaultKubernetesAuthMount].(string), args[FlagPsatPath].(string)); err != nil {
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
