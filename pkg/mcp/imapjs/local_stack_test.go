package imapjs

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
)

type localTokenClaims struct {
	Issuer            string `json:"iss"`
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
}

func TestExecuteIMAPJSAgainstLocalKeycloakAndDovecot(t *testing.T) {
	if os.Getenv("SMAILNAIL_LOCAL_STACK_TEST") != "1" {
		t.Skip("set SMAILNAIL_LOCAL_STACK_TEST=1 to run the local Keycloak + Dovecot integration test")
	}

	ctx := context.Background()
	baseURL := envOrDefault("SMAILNAIL_LOCAL_KEYCLOAK_URL", "http://127.0.0.1:18080")
	realm := envOrDefault("SMAILNAIL_LOCAL_KEYCLOAK_REALM", "smailnail-dev")
	clientID := envOrDefault("SMAILNAIL_LOCAL_KEYCLOAK_CLIENT_ID", "smailnail-mcp")
	username := envOrDefault("SMAILNAIL_LOCAL_KEYCLOAK_USERNAME", "alice")
	password := envOrDefault("SMAILNAIL_LOCAL_KEYCLOAK_PASSWORD", "secret")

	if err := ensureLocalKeycloakPasswordGrantSetup(ctx, baseURL, realm, clientID, username, password); err != nil {
		t.Fatalf("ensureLocalKeycloakPasswordGrantSetup() error = %v", err)
	}

	token, err := fetchPasswordGrantToken(ctx, baseURL+"/realms/"+realm+"/protocol/openid-connect/token", clientID, username, password)
	if err != nil {
		t.Fatalf("fetchPasswordGrantToken() error = %v", err)
	}
	claims, err := parseLocalTokenClaims(token)
	if err != nil {
		t.Fatalf("parseLocalTokenClaims() error = %v", err)
	}

	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()
	if err := hostedapp.BootstrapApplicationDB(ctx, db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	secretConfig, err := secrets.LoadConfigFromSettings(&secrets.Settings{
		KeyBase64: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
		KeyID:     "test-key",
	})
	if err != nil {
		t.Fatalf("LoadConfigFromSettings() error = %v", err)
	}

	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo)
	resolved, err := identityService.ResolveOrProvisionUser(ctx, identity.ExternalPrincipal{
		Issuer:            claims.Issuer,
		Subject:           claims.Subject,
		ProviderKind:      identity.ProviderKindOIDC,
		ClientID:          clientID,
		Email:             claims.Email,
		EmailVerified:     claims.EmailVerified,
		PreferredUsername: claims.PreferredUsername,
		DisplayName:       claims.Name,
	})
	if err != nil {
		t.Fatalf("ResolveOrProvisionUser() error = %v", err)
	}

	accountService := accounts.NewService(accounts.NewRepository(db), secretConfig)
	account, err := accountService.Create(ctx, resolved.User.ID, accounts.CreateInput{
		Label:          "Local Dovecot",
		Server:         envOrDefault("SMAILNAIL_LOCAL_DOVECOT_HOST", "127.0.0.1"),
		Port:           envOrDefaultInt("SMAILNAIL_LOCAL_DOVECOT_PORT", 993),
		Username:       envOrDefault("SMAILNAIL_LOCAL_DOVECOT_USERNAME", "a"),
		Password:       envOrDefault("SMAILNAIL_LOCAL_DOVECOT_PASSWORD", "pass"),
		MailboxDefault: envOrDefault("SMAILNAIL_LOCAL_DOVECOT_MAILBOX", "INBOX"),
		Insecure:       envOrDefault("SMAILNAIL_LOCAL_DOVECOT_INSECURE", "1") == "1",
		AuthKind:       accounts.AuthKindPassword,
		MCPEnabled:     true,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	runtime := newSharedIdentityRuntimeWithServices(db, identityService, accountService)
	result, err := runtime.middleware()(executeIMAPJSHandler)(
		embeddable.WithAuthPrincipal(context.Background(), embeddable.AuthPrincipal{
			Issuer:            claims.Issuer,
			Subject:           claims.Subject,
			ClientID:          clientID,
			Email:             claims.Email,
			EmailVerified:     claims.EmailVerified,
			PreferredUsername: claims.PreferredUsername,
			DisplayName:       claims.Name,
		}),
		map[string]interface{}{
			"code": `
const smailnail = require("smailnail");
const svc = smailnail.newService();
const session = svc.connect({ accountId: "` + account.ID + `" });
const result = { mailbox: session.mailbox };
session.close();
result;
`,
		},
	)
	if err != nil {
		t.Fatalf("executeIMAPJSHandler error = %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got %#v", result)
	}

	var decoded ExecuteIMAPJSResponse
	if err := json.Unmarshal([]byte(result.Content[0].Text), &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	value, ok := decoded.Value.(map[string]interface{})
	if !ok {
		t.Fatalf("decoded.Value type = %T", decoded.Value)
	}
	if value["mailbox"] != envOrDefault("SMAILNAIL_LOCAL_DOVECOT_MAILBOX", "INBOX") {
		t.Fatalf("unexpected mailbox payload: %#v", value)
	}
}

func ensureLocalKeycloakPasswordGrantSetup(ctx context.Context, baseURL, realm, clientID, username, password string) error {
	adminToken, err := fetchPasswordGrantToken(ctx, baseURL+"/realms/master/protocol/openid-connect/token", "admin-cli", "admin", "admin")
	if err != nil {
		return err
	}

	clientRep, err := getLocalKeycloakClient(ctx, baseURL, realm, clientID, adminToken)
	if err != nil {
		return err
	}
	clientRep["directAccessGrantsEnabled"] = true
	if err := putLocalKeycloakClient(ctx, baseURL, realm, fmt.Sprint(clientRep["id"]), adminToken, clientRep); err != nil {
		return err
	}

	userID, err := ensureLocalKeycloakUser(ctx, baseURL, realm, username, password, adminToken)
	if err != nil {
		return err
	}
	if userID == "" {
		return fmt.Errorf("keycloak user id is empty")
	}
	return nil
}

func getLocalKeycloakClient(ctx context.Context, baseURL, realm, clientID, adminToken string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/admin/realms/"+realm+"/clients?clientId="+url.QueryEscape(clientID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("client lookup returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var clients []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, err
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("client %q not found", clientID)
	}
	return clients[0], nil
}

func putLocalKeycloakClient(ctx context.Context, baseURL, realm, clientUUID, adminToken string, rep map[string]interface{}) error {
	body, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, baseURL+"/admin/realms/"+realm+"/clients/"+clientUUID, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("client update returned %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return nil
}

func ensureLocalKeycloakUser(ctx context.Context, baseURL, realm, username, password, adminToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/admin/realms/"+realm+"/users?username="+url.QueryEscape(username), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("user lookup returned %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", err
	}

	userID := ""
	if len(users) > 0 {
		userID = fmt.Sprint(users[0]["id"])
	} else {
		createBody, _ := json.Marshal(map[string]interface{}{
			"username": username,
			"email":    username + "@example.com",
			"enabled":  true,
		})
		createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/admin/realms/"+realm+"/users", bytes.NewReader(createBody))
		if err != nil {
			return "", err
		}
		createReq.Header.Set("Authorization", "Bearer "+adminToken)
		createReq.Header.Set("Content-Type", "application/json")

		createResp, err := http.DefaultClient.Do(createReq)
		if err != nil {
			return "", err
		}
		defer createResp.Body.Close()
		if createResp.StatusCode != http.StatusCreated {
			data, _ := io.ReadAll(createResp.Body)
			return "", fmt.Errorf("user create returned %d: %s", createResp.StatusCode, strings.TrimSpace(string(data)))
		}

		location := createResp.Header.Get("Location")
		parts := strings.Split(strings.TrimRight(location, "/"), "/")
		userID = parts[len(parts)-1]
	}

	passwordBody, _ := json.Marshal(map[string]interface{}{
		"type":      "password",
		"value":     password,
		"temporary": false,
	})
	updateBody, _ := json.Marshal(map[string]interface{}{
		"username":        username,
		"email":           username + "@example.com",
		"firstName":       "Alice",
		"lastName":        "Example",
		"enabled":         true,
		"emailVerified":   true,
		"requiredActions": []string{},
	})
	updateReq, err := http.NewRequestWithContext(ctx, http.MethodPut, baseURL+"/admin/realms/"+realm+"/users/"+userID, bytes.NewReader(updateBody))
	if err != nil {
		return "", err
	}
	updateReq.Header.Set("Authorization", "Bearer "+adminToken)
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		return "", err
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode != http.StatusNoContent {
		data, _ := io.ReadAll(updateResp.Body)
		return "", fmt.Errorf("user update returned %d: %s", updateResp.StatusCode, strings.TrimSpace(string(data)))
	}

	passwordReq, err := http.NewRequestWithContext(ctx, http.MethodPut, baseURL+"/admin/realms/"+realm+"/users/"+userID+"/reset-password", bytes.NewReader(passwordBody))
	if err != nil {
		return "", err
	}
	passwordReq.Header.Set("Authorization", "Bearer "+adminToken)
	passwordReq.Header.Set("Content-Type", "application/json")

	passwordResp, err := http.DefaultClient.Do(passwordReq)
	if err != nil {
		return "", err
	}
	defer passwordResp.Body.Close()
	if passwordResp.StatusCode != http.StatusNoContent {
		data, _ := io.ReadAll(passwordResp.Body)
		return "", fmt.Errorf("password reset returned %d: %s", passwordResp.StatusCode, strings.TrimSpace(string(data)))
	}

	return userID, nil
}

func fetchPasswordGrantToken(ctx context.Context, tokenURL, clientID, username, password string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", clientID)
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", err
	}
	if decoded.AccessToken == "" {
		return "", fmt.Errorf("access token missing in response")
	}
	return decoded.AccessToken, nil
}

func parseLocalTokenClaims(token string) (*localTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid JWT format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims localTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var parsed int
		if _, err := fmt.Sscanf(value, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return fallback
}
