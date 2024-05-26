// cmd/root_test.go
package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/apis/clientauthentication"
)

func TestFederatePsatSubcmd(t *testing.T) {
	// Create a buffer to capture command output
	buf := new(bytes.Buffer)

	// create root command
	cmd := New(true)

	// Set the output of the command to the buffer
	cmd.SetOut(buf)

	// Execute the command with no arguments
	cmd.SetArgs([]string{"federate", "psat"})
	err := cmd.Execute()

	expectedError := `required flag(s) "vault-kubernetes-auth-mount" not set`
	// expectedPattern := "required flag(s).*not set"
	assert.EqualError(t, err, expectedError)
	// assert.Regexp(t, regexp.MustCompile(expectedPattern), err)
}

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	originalStdout := os.Stdout
	os.Stdout = w

	out := make(chan string)
	go func() {
		var buf strings.Builder
		io.Copy(&buf, r)
		out <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = originalStdout
	return <-out
}

func TestFederatePsatEndToEnd(t *testing.T) {
	var execCredential clientauthentication.ExecCredential
	os.Setenv("KUBERNETES_EXEC_INFO", `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1","spec":{"cluster":{"server":"https://k8s.example.com","config":null},"interactive":false}}`)

	output := captureOutput(func() {
		// create root command
		cmd := New(true)
		// Execute the command with arguments
		cmd.SetArgs([]string{"federate", "psat",
			"--vault-kubernetes-auth-mount=/kubernetes/argocd",
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
