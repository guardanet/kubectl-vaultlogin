package errors

// implements kubectl-vaultlogin error type to be able to differentiate between errors from cobra command and the kubectl-vaultlogin plugin
type KvlError struct {
	Message string
}

// KvlError is a Error type
func (e KvlError) Error() string {
	return e.Message
}

func New(message string) error {
	return KvlError{Message: message}
}
