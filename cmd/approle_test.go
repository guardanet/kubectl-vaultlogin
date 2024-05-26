// cmd/root_test.go
package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/apis/clientauthentication"
)

func TestFederateApproleSubcmd(t *testing.T) {
	// Create a buffer to capture command output
	buf := new(bytes.Buffer)

	// create root command
	cmd := New(true)

	// Set the output of the command to the buffer
	cmd.SetOut(buf)

	// Execute the command with no arguments
	cmd.SetArgs([]string{"federate", "approle"})
	err := cmd.Execute()

	expectedError := `GetExecCredentialFromEnv(): kubectl-vaultlogin is a kubectl credential plugin and requires an ExecCredetnial to be provided in the KUBERNETES_EXEC_INFO env variable. Exiting as the variable is unset or empty`
	assert.EqualError(t, err, expectedError)
	// assert.EqualError(t, err, expectedError)
	// assert.Regexp(t, regexp.MustCompile(expectedError), err.Error())
}

func TestFederateApproleEndToEnd(t *testing.T) {
	var execCredential clientauthentication.ExecCredential
	// KUBERNETES_EXEC_INFO=$(cat input-ec-nocluster) ./kubectl-vaultlogin federate psat -a /kubernetes/c1 -p /home/kwaberski/go/src/kvl/kvl-token -c c4 -v https://vault.corp.guardanet.net:8200
	os.Setenv("KUBERNETES_EXEC_INFO", `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1","spec":{"cluster":{"server":"https://k8s.example.com","config":null},"interactive":false}}`)
	// os.Setenv("APPROLE_ROLE_ID", "d162568c-d3c2-daee-235b-467ff1cd74e2")
	// os.Setenv("APPROLE_SECRET_ID", "731d4403-55e1-e87a-3425-ab4eb9ebb337")
	output := captureOutput(func() {
		// create root command
		cmd := New(true)
		// Execute the command with arguments
		cmd.SetArgs([]string{"federate", "approle",
			"--vault-address=https://vault.corp.guardanet.net:8200",
		})
		err := cmd.Execute()
		assert.NoError(t, err)
	})
	// let's unmarshal what we received to an ExecCredetnial struct
	err := json.Unmarshal([]byte(output), &execCredential)
	// if the unmarshal operation failed or Token is empty, then throw an error
	if err != nil || execCredential.Status.Token == "" {
		t.Error("Didn't receive a valid ExecCredential")
	}
}
