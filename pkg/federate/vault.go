package federate

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	vaultcg "github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

// authToVaultWithApprole authenticates to Vault using approle and retrieving RoleId and SecretId from env variables APPROLE_ROLE_ID and APPROLE_SECRET_ID respectively.
// Upon successful authentication it popules the client with the recevied vault token.
func authToVaultWithApprole(ctx *context.Context, client *vaultcg.Client, mountPath string) error {
	resp, err := client.Auth.AppRoleLogin(
		*ctx,
		schema.AppRoleLoginRequest{
			RoleId:   os.Getenv("APPROLE_ROLE_ID"),
			SecretId: os.Getenv("APPROLE_SECRET_ID"),
		},
		vaultcg.WithMountPath(mountPath), // optional, defaults to "approle"
	)
	if err != nil {
		return fmt.Errorf("authToVaultWithApprole() AppRoleLogin: %s", err)
	}
	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		return fmt.Errorf("authToVaultWithApprole() SetToken: %s", err)
	}
	return nil
}

// authToVaultWithKubernetes authenticates to Vault using kubernetes authentication and exchanging its PSAT for a vault token.
// Upon successful authentication it popules the client with the recevied vault token.
func authToVaultWithKubernetes(ctx *context.Context, client *vaultcg.Client, vaultKubernetesLoginRole string, mountPath string, tokenPath string) error {
	path, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Auth.KubernetesLogin(
		*ctx,
		schema.KubernetesLoginRequest{
			// Jwt:  os.Getenv("KVL_TOKEN"),
			Jwt:  string(path),
			Role: vaultKubernetesLoginRole},
		vaultcg.WithMountPath(mountPath),
	)
	if err != nil {
		return fmt.Errorf("authToVaultWithKubernetes() KubernetesLogin: %s", err)
	}
	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		return fmt.Errorf("authToVaultWithKubernetes() SetToken: %s", err)
	}
	return nil
}

// generateK8sToken returns a kubernetes bearer token that it obtained from Vault
func generateK8sToken(ctx *context.Context, client *vaultcg.Client, roleName string, clusterName string) (string, error) {

	resp, err := client.Secrets.KubernetesGenerateCredentials(*ctx, roleName, schema.KubernetesGenerateCredentialsRequest{KubernetesNamespace: "kube-priv"},
		vaultcg.WithMountPath("/kubernetes/"+clusterName),
	)
	if err != nil {
		return "", fmt.Errorf("generateK8sToken() KubernetesGenerateCredentials: cluster=%s, role=%s, error=%s", clusterName, roleName, err)
	}
	return resp.Data["service_account_token"].(string), nil
}

func generateFakeK8sToken() string {
	// Define the signing method and the secret key
	signingKey := []byte("fake_secret_key")

	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "1234567890",                            // subject
		"name": "John Doe",                              // name
		"iat":  time.Now().Unix(),                       // issued at
		"exp":  time.Now().Add(minTokenDuration).Unix(), // expiration time
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		fmt.Println("Error signing token:", err)
		return ""
	}
	return tokenString
}
