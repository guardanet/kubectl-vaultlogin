package federate

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthentication "k8s.io/client-go/pkg/apis/clientauthentication/v1"
)

var mockArgs = map[string]any{
	FlagClusterName:  "dev",
	FlagVaultAddress: "https://localhost:8200",
}
var mockArgsNocluster = map[string]any{
	FlagClusterName:  "",
	FlagVaultAddress: "https://localhost:8200",
}

// tests if error is reported when KUBERNETES_EXEC_INFO is not set
func TestPrepFederationNoExec(t *testing.T) {
	err := prepFederation(&mockArgs)

	expectedError := `GetExecCredentialFromEnv(): kubectl-vaultlogin is a kubectl credential plugin and requires an ExecCredetnial to be provided in the KUBERNETES_EXEC_INFO env variable. Exiting as the variable is unset or empty`
	assert.EqualError(t, err, expectedError)

}

// tests if error is reported when KUBERNETES_EXEC_INFO holds a valid ExecCredential
// but without Spec.Cluster and cluster-name flag isn't set
func TestPrepFederationExecnospecclusterNoclustername(t *testing.T) {
	os.Setenv("KUBERNETES_EXEC_INFO", `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1","spec":{"interactive":false}}`)
	err := prepFederation(&mockArgsNocluster)

	expectedPattern := "and cluster-name flag is unset or empty"
	assert.Regexp(t, regexp.MustCompile(expectedPattern), err)

}

// tests if no error is reported and cluster name is set to the value provided in the cluster name flag
// when KUBERNETES_EXEC_INFO holds a valid ExecCredential but without Spec.Cluster
func TestPrepFederationExecnospecclusterClustername(t *testing.T) {
	os.Setenv("KUBERNETES_EXEC_INFO", `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1","spec":{"interactive":false}}`)
	cname := mockArgs[FlagClusterName].(string)
	err := prepFederation(&mockArgs)
	assert.NoError(t, err)
	assert.Equal(t, cname, mockArgs[FlagClusterName].(string))
}

// tests if no error is reported and cluster name is set to the value of Spec.Cluster
// when cluster name flag is set to a value and at the same time
// KUBERNETES_EXEC_INFO holds a valid ExecCredential with Spec.Cluster and Server value specified
func TestPrepFederationExecClustername(t *testing.T) {
	os.Setenv("KUBERNETES_EXEC_INFO", `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1","spec":{"cluster":{"server":"https://k8s.example.com:443","config":null},"interactive":false}}`)
	cname := "k8s"
	err := prepFederation(&mockArgs)
	assert.NoError(t, err)
	assert.Equal(t, cname, mockArgs[FlagClusterName].(string))
}

// tests if no error is reported and cluster name is set to the value of Spec.Cluster
// when cluster name flag isn't set
// KUBERNETES_EXEC_INFO holds a valid ExecCredential with Spec.Cluster and Server value specified
func TestPrepFederationExecNoclustername(t *testing.T) {
	os.Setenv("KUBERNETES_EXEC_INFO", `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1","spec":{"cluster":{"server":"https://k8s.example.com:443","config":null},"interactive":false}}`)
	cname := "k8s"
	err := prepFederation(&mockArgsNocluster)
	assert.NoError(t, err)
	assert.Equal(t, cname, mockArgs[FlagClusterName].(string))
}

// tests if no error is reported when setting up a Vault client
func TestPrepVaultClient(t *testing.T) {
	err := prepVaultClient(mockArgs[FlagVaultAddress].(string))
	assert.NoError(t, err)
}

// test if error is reported when improper path is supplied as vault-kubernetes-auth-mount
func TestPrepVaultAuthMountWrong(t *testing.T) {
	mockArgs := map[string]any{
		"vault-kubernetes-auth-mount": "wrong",
	}
	err := prepVaultAuthMount(&mockArgs)
	expectedPattern := "malformed vault authentication mount path"
	assert.Regexp(t, regexp.MustCompile(expectedPattern), err)
}

// test if no error is reported when a correct path is supplied as vault-kubernetes-auth-mount

func TestPrepVaultAuthMount(t *testing.T) {
	mockArgs := map[string]any{
		"vault-kubernetes-auth-mount": "/kubernetes/argocd",
	}
	err := prepVaultAuthMount(&mockArgs)
	assert.NoError(t, err)
}

func TestGetDownstreamClusterName(t *testing.T) {
	mockExecCred := clientauthentication.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ExecCredential",
			APIVersion: "client.authentication.k8s.io/v1",
		},
		Spec: clientauthentication.ExecCredentialSpec{
			Cluster: &clientauthentication.Cluster{
				Server: "https://k8s.example.com:443",
			},
			Interactive: false,
		},
	}

	cname, err := getDownstreamClusterName(&mockExecCred)

	assert.NoError(t, err)
	assert.Equal(t, "k8s", cname)

}
