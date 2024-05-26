package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const kvlTestError = "Test kubernetes-vaultlogin error"

func TestNew(t *testing.T) {
	kvlErr := New(kvlTestError)
	errType := fmt.Sprintf("%T", kvlErr)
	assert.Equal(t, "errors.KvlError", errType)
}

func TestError(t *testing.T) {
	kvlerr := New(kvlTestError)
	errMessage := kvlerr.Error()
	assert.Equal(t, kvlTestError, errMessage)
}
