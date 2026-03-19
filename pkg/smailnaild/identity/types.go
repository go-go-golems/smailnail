package identity

import "time"

const ProviderKindOIDC = "oidc"

type User struct {
	ID           string    `db:"id" json:"id"`
	PrimaryEmail string    `db:"primary_email" json:"primaryEmail"`
	DisplayName  string    `db:"display_name" json:"displayName"`
	AvatarURL    string    `db:"avatar_url" json:"avatarUrl"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

type ExternalIdentity struct {
	ID                string    `db:"id" json:"id"`
	UserID            string    `db:"user_id" json:"userId"`
	Issuer            string    `db:"issuer" json:"issuer"`
	Subject           string    `db:"subject" json:"subject"`
	ProviderKind      string    `db:"provider_kind" json:"providerKind"`
	Email             string    `db:"email" json:"email"`
	EmailVerified     bool      `db:"email_verified" json:"emailVerified"`
	PreferredUsername string    `db:"preferred_username" json:"preferredUsername"`
	RawClaimsJSON     string    `db:"raw_claims_json" json:"-"`
	CreatedAt         time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time `db:"updated_at" json:"updatedAt"`
}

type WebSession struct {
	ID         string    `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"userId"`
	Issuer     string    `db:"issuer" json:"issuer"`
	Subject    string    `db:"subject" json:"subject"`
	ExpiresAt  time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	LastSeenAt time.Time `db:"last_seen_at" json:"lastSeenAt"`
}

type ExternalPrincipal struct {
	Issuer            string         `json:"issuer"`
	Subject           string         `json:"subject"`
	ProviderKind      string         `json:"providerKind"`
	ClientID          string         `json:"clientId,omitempty"`
	Email             string         `json:"email,omitempty"`
	EmailVerified     bool           `json:"emailVerified"`
	PreferredUsername string         `json:"preferredUsername,omitempty"`
	DisplayName       string         `json:"displayName,omitempty"`
	AvatarURL         string         `json:"avatarUrl,omitempty"`
	Scopes            []string       `json:"scopes,omitempty"`
	Claims            map[string]any `json:"claims,omitempty"`
}

type ResolvedIdentity struct {
	User             *User             `json:"user"`
	ExternalIdentity *ExternalIdentity `json:"externalIdentity"`
}
