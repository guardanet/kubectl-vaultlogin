// cmd/root_test.go
package cmd

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	// Create a buffer to capture command output
	buf := new(bytes.Buffer)

	// create root command
	cmd := New(true)

	// Set the output of the command to the buffer
	cmd.SetOut(buf)

	// Execute the command with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	// Check for errors
	assert.NoError(t, err)

	// Check the output
	output := buf.String()
	expectedPattern := "(?s)kubectl-vaultlogin is a client-go credential plugin for kubectl.*federate.*version.*--cluster-name.*--vault-address"
	// assert.Equal(t, expectedOutput, output)
	assert.Regexp(t, regexp.MustCompile(expectedPattern), output)
}
