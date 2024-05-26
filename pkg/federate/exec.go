package federate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthentication "k8s.io/client-go/pkg/apis/clientauthentication/v1"
)

// GetExecCredentialFromEnv populates ExecCredential with what it received from kubectl in the KUBERNETES_EXEC_INFO env var
func getExecCredentialFromEnv() (*clientauthentication.ExecCredential, error) {
	var execCredential clientauthentication.ExecCredential

	env := os.Getenv(execInfoEnv)
	//env := string('{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1beta1","spec":{"cluster":{"server":"https://c1:443","certificate-authority-data":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJek1EUXlOakV3TXpreU1sb1hEVE16TURReU16RXdNemt5TWxvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTDZECnVLWmRPVG9JQTgwelZoTU5ieFg3MGR5WnVQSGd0WHBMbms4VFlYb1FYSEVQTnIyZkE1cGFORmFqOCtOc3JWb3kKWlQybXFBNVkrTkRhS2Q0bVczUU93alJvQ1ErazFsVVpwV2hMakVxY0RuVy90Vnd3dmJYQkVMMVVBUmlVaC9rSgpXelEzMEZST09aUTh1OGpySnk3REVZUCt4TmpnR0xIZ3dVdFAweVNlMUJaOVF6amRqalN4UDZFdnlLUy94Z2xICm5DbFROUnVrMXlLNWJNL1VlcnFQSWs0U3hEME1qSHdPbkhlRzBCNUNMK08zZ3hGVlIrbVZwV3lGVmIyY21XU1EKUkxyOEhsdzg0aUxlaTRjVGcrN1VDNnpSWVQxOGhGdnZxYWpoZlEwQUF4MXJ3TGR0cUJQcHlMK1dvWFQ3bGNMegpuNGg1Y0ZUeXlPTGlKMTdXeHhNQ0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZBeVVrNUlNUG5TVTFhaEE2R1hJMnI0YVVmTVRNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBQmlkbVpsWXJ2eWoxWk5MVmttUwpENStYdGRSREZyY1p2ZXdGUTNTNHFjSFgvUGFYWFE3R3FCYUtpSWQvS2tMcGJVSlJXdFFyOXhCajhyYXRwUTB6Cm1adTVTQWNXVFBwcGo0K0VtcmpUZzBGWXU1bGM1T1lqUmRoRGUwcVV2RGFrUW9GeFhOVW4zaDZCb0JpM1FiV0UKM29Dd0lwTXg5aXMvc2JtcUttbEx1U01pWW5sUE5ZQllHczFUUXhkSmdEU3grS3FrZG9hOHMxZG1hdE83M1BIcQpXK1ZOL3dTZGZTbUUxQmV1VndOMG5uTnZGeXVkSzBCeXlvTVVldjlFV2dOOEVtd09Ya0RXODVtKytoNHVpOXZECjgxb2gyOUxEODFpZlVBTGRCR2VmV0FCOW1YUTQvYlVWbGM0NFVkU1lIL0N2dHEydzRmbEdDVGNNL2E4VGJEMDkKb3hBPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==","config":null},"interactive":true}}')
	if env == "" {
		return nil, errors.New("GetExecCredentialFromEnv(): kubectl-vaultlogin is a kubectl credential plugin and requires an ExecCredetnial to be provided in the KUBERNETES_EXEC_INFO env variable. Exiting as the variable is unset or empty")
	}

	if err := json.Unmarshal([]byte(env), &execCredential); err != nil {
		return nil, fmt.Errorf("GetExecCredentialFromEnv(): cannot unmarshal %q to ExecCredential: %w", env, err)
	}
	return &execCredential, nil
}

// printExecCredential prints ExecCredential struct in JSON format
func printExecCredential(w io.Writer, execCredentialPointer *clientauthentication.ExecCredential) error {
	// marshal execCredentialPointer struct to JSON
	data, err := json.Marshal(*execCredentialPointer)
	if err != nil {
		return fmt.Errorf("printExecCredential: cannot marshal ExecCredential to JSON: %w", err)
	}

	// Write JSON to io.Writer (could be file or STDOUT)
	_, err = fmt.Fprintf(w, "%s\n", data)
	if err != nil {
		return fmt.Errorf("printExecCredential: could not write JSON to %T: %w", w, err)
	}
	return nil
}

// assembleExecCredential prepares ExecCredential by adding ExecCredentialStatus with a bearer token and expiration
func assembleExecCredential(execCredentialPointer *clientauthentication.ExecCredential, token string) error {
	// take a pointer to the received ExecCredential and add .Status field to it containing the received k8s token
	// var execCredentialStatus clientauthentication.ExecCredentialStatus
	// execCredentialStatus.Token = token
	// (*execCredentialPointer).Status = &execCredentialStatus
	// return nil
	(*execCredentialPointer).Status = &clientauthentication.ExecCredentialStatus{
		ExpirationTimestamp: &metav1.Time{Time: time.Now().Add(tokenDuration)},
		Token:               token,
	}
	return nil
}
