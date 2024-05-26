package federate

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	clientauthentication "k8s.io/client-go/pkg/apis/clientauthentication/v1"

	vaultcg "github.com/hashicorp/vault-client-go"
)

// const to define cobra command flag name that supplies vault address
const FlagVaultAddress = "vault-address"

// const to define cobra command flag name that supplies downstream cluster name
const FlagClusterName = "cluster-name"

// tokenDuration represents expiration set in the ExecCredentialStatus
var tokenDuration time.Duration

// minTokenDuration establishes a 15 min minimum allowed expiration for ExecCredentialStatus
const minTokenDuration time.Duration = time.Minute * 15

// vaultKubernetesLoginRole represents a role in Vault's kuberentes authentication backend used to authetinticate with k8s sourced PSAT and subsequently obtain a Vault token
var vaultKubernetesLoginRole string

// vaultK8sSecretRole represents a role in Vault's kuberentes secret backend used to obtain k8sToken
var vaultK8sSecretRole string

// execInfoEnv defines a variable in which ExecCredential is passed
const execInfoEnv = "KUBERNETES_EXEC_INFO"

// hashicorp vault client
var client *vaultcg.Client

// pointer to received ExecCredential
var inputExecCredentialPtr *clientauthentication.ExecCredential

// kubernetes bearer token
var k8sToken string

// prepFederation performs the following preparation tasks:
// 1. calls setVariables
// 2. captures ExecCredential from KUBERNETES_EXEC_INFO
// 3. prepares a vault client,
// 4. checks if ExecCredential.Spec.Server exists and if it doesn't then uses the value supplied as part of cluser-name flag
func prepFederation(args *map[string]any) error {
	// func prepFederation(args *map[string]any) {

	var err error

	// set variables for vault kubernetes authentication and secret roles as well as token duration
	setVariables()

	// capture received ExecCredential
	inputExecCredentialPtr, err = getExecCredentialFromEnv()
	if err != nil {
		return err
	}

	// We need to ensure that cluster name is set to inputExecCredentialPtr.Spec.Cluster.Server or to the value of the cluster-name flag in that order of preference
	if cname, err := getDownstreamClusterName(inputExecCredentialPtr); err != nil {
		if (*args)[FlagClusterName].(string) == "" {
			return fmt.Errorf("%s and cluster-name flag is unset or empty", err)
		} else {
			if !isValidHostname((*args)[FlagClusterName].(string)) {
				return fmt.Errorf("cluster-name must be a string that is a valid dns name: %s", (*args)[FlagClusterName].(string))
			}
		}
	} else {
		if !isValidHostname(cname) {
			return fmt.Errorf("cluster-name must be an alphanumeric string: %s", cname)
		}
		(*args)[FlagClusterName] = cname
	}
	return nil
}

func prepVaultClient(vaddr string) error {
	var err error
	client, err = vaultcg.New(
		vaultcg.WithAddress(vaddr),
		vaultcg.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failure preparing vault client: %s", err)
	}
	return nil
}

// prepVaultAuthMount prepares a Vault authentication mount point only when kuberenetes authentication is used.
// this function is not necessary when approle authentication is used as default value for the mount point is assigned
func prepVaultAuthMount(args *map[string]any) error {

	// first if the vaultAuthMount was not provided then assign it a value of the VAULT_AUTH_MOUNT env variable.
	// we dont check if the env variable exists because right after we check it the vaultAuthMount a value that follows an expected pattern
	if (*args)[FlagVaultKubernetesAuthMount].(string) == "" {
		(*args)[FlagVaultKubernetesAuthMount] = os.Getenv("VAULT_AUTH_MOUNT")
	}
	pattern := "^/kubernetes.*$"
	regex := regexp.MustCompile(pattern)
	if !regex.MatchString((*args)[FlagVaultKubernetesAuthMount].(string)) {
		return fmt.Errorf("malformed vault authentication mount path: %s. The value should be in the form of /kubernetes*", (*args)[FlagVaultKubernetesAuthMount].(string))
	}
	return nil
}

// Applies either values from environment variables or default to the following variables:
// vaultKubernetesLoginRole, vaultK8sSecretRole, tokenDuration
func setVariables() {
	var (
		err    error
		exists bool
	)

	// Get vaultKubernetesLoginRole from an env variable or apply a default
	if vaultKubernetesLoginRole, exists = os.LookupEnv("VAULT_K8S_LOGIN_ROLE"); !exists || vaultKubernetesLoginRole == "" {
		vaultKubernetesLoginRole = "kvl-login"
	}

	// Get vaultK8sSecretRole from an env variable or apply a default
	if vaultK8sSecretRole, exists = os.LookupEnv("VAULT_K8S_SECRET_ROLE"); !exists || vaultK8sSecretRole == "" {
		vaultK8sSecretRole = "kvl-edit-role"
	}

	// Get tokenDuration from an env variable or apply a minTokenDuration. ExecCredenatialStatus.Expiration > Min vault TTL (10mins)
	if value, exists := os.LookupEnv("TOKEN_DURATION"); exists {
		tokenDuration, err = time.ParseDuration(value)
		if err != nil {
			log.Fatal(err)
		}
		if tokenDuration < minTokenDuration {
			tokenDuration = minTokenDuration
		}
	} else {
		tokenDuration = minTokenDuration
	}
}

// getDownstreamClusterName extracts downstream cluster name from inputExecCredentialPtr.Spec.Cluster.Server and in the case when it is not supplied
// it assigns it a value passed by the downstreamClusterName command lime flag.
func getDownstreamClusterName(execCredentialPointer *clientauthentication.ExecCredential) (string, error) {
	// first check if after JSON unmarshall executed by the GetExecCredentialFromEnv() function
	// the Cluster is indeed part of the ExecCredential.
	// SUPER IMPORTANT: if it is not then because cluster is a pointer it will be nil
	if execCredentialPointer.Spec.Cluster == nil {
		return "", errors.New("getDownstreamClusterName(): cluster info - ExecCredential.Spec.Cluster - not provided as part of execCredential")
	} else {
		// get downstream cluster URL from ExecCredential
		// wee need a hostname part from the url, we parse the URL and then url.Hostname() is what we need
		url, err := url.Parse(execCredentialPointer.Spec.Cluster.Server)
		if err != nil {
			return "", fmt.Errorf("getDownstreamClusterName(): error when retrieving the cluster url from execcredential: %s", err)
		}
		// now we have the FQDN, lets extract the hostname
		parts := strings.Split(url.Hostname(), ".")
		return parts[0], nil
	}
}

// isValidURL checks if URL is valid, uses https and doesn't include relative paths or queries
func isValidURL(addr string) error {
	url, err := url.ParseRequestURI(addr)
	if err != nil {
		return fmt.Errorf("mal formed vault-address: %s", addr)
	} else {
		if url.Scheme != "https" {
			return fmt.Errorf("only https is allowed in vault-address: %s", addr)
		}
		if url.Path != "" {
			return fmt.Errorf("relative paths are not allowed in vault-address: %s", addr)
		}
		if url.RawQuery != "" {
			return fmt.Errorf("queries are not allowed in vault-address: %s", addr)
		}
	}
	return nil
}

// isAlphanumeric check id a string is alphanumeric and can also take into cosideration a sepacial separator
// for checking if strings resemble a vault path ex. /path/subpath/subsubpath
func isAlphanumeric(s string) bool {
	// Define a regular expression that matches only alphanumeric characters
	pattern := `^[a-zA-Z0-9]+$`
	var alphanumericRegex = regexp.MustCompile(pattern)
	return alphanumericRegex.MatchString(s)
}

func isAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

func isValidHostname(hostname string) bool {
	// Regular expression for a valid DNS hostname
	var hostnameRegexp = regexp.MustCompile(`^(?i)[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*\.?$`)

	// Check if the hostname matches the regular expression
	if !hostnameRegexp.MatchString(hostname) {
		return false
	}

	// Ensure each label of the hostname is not longer than 63 characters
	labels := regexp.MustCompile(`\.`).Split(hostname, -1)
	for _, label := range labels {
		if len(label) > 63 {
			return false
		}
	}

	// Ensure the entire hostname is not longer than 253 characters
	if len(hostname) > 253 {
		return false
	}

	return true
}
