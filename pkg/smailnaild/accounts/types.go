package accounts

import "time"

type Account struct {
	ID               string    `db:"id" json:"id"`
	UserID           string    `db:"user_id" json:"userId"`
	Label            string    `db:"label" json:"label"`
	ProviderHint     string    `db:"provider_hint" json:"providerHint"`
	Server           string    `db:"server" json:"server"`
	Port             int       `db:"port" json:"port"`
	Username         string    `db:"username" json:"username"`
	MailboxDefault   string    `db:"mailbox_default" json:"mailboxDefault"`
	Insecure         bool      `db:"insecure" json:"insecure"`
	AuthKind         string    `db:"auth_kind" json:"authKind"`
	SecretCiphertext string    `db:"secret_ciphertext" json:"-"`
	SecretNonce      string    `db:"secret_nonce" json:"-"`
	SecretKeyID      string    `db:"secret_key_id" json:"secretKeyId"`
	IsDefault        bool      `db:"is_default" json:"isDefault"`
	MCPEnabled       bool      `db:"mcp_enabled" json:"mcpEnabled"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time `db:"updated_at" json:"updatedAt"`
}

type AccountTest struct {
	ID              string    `db:"id" json:"id"`
	IMAPAccountID   string    `db:"imap_account_id" json:"imapAccountId"`
	TestMode        string    `db:"test_mode" json:"testMode"`
	Success         bool      `db:"success" json:"success"`
	TCPOK           bool      `db:"tcp_ok" json:"tcpOk"`
	LoginOK         bool      `db:"login_ok" json:"loginOk"`
	MailboxSelectOK bool      `db:"mailbox_select_ok" json:"mailboxSelectOk"`
	ListOK          bool      `db:"list_ok" json:"listOk"`
	SampleFetchOK   bool      `db:"sample_fetch_ok" json:"sampleFetchOk"`
	WriteProbeOK    *bool     `db:"write_probe_ok" json:"writeProbeOk,omitempty"`
	WarningCode     string    `db:"warning_code" json:"warningCode,omitempty"`
	ErrorCode       string    `db:"error_code" json:"errorCode,omitempty"`
	ErrorMessage    string    `db:"error_message" json:"errorMessage,omitempty"`
	DetailsJSON     string    `db:"details_json" json:"detailsJson"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
}
