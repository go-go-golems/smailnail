package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidPrincipal = errors.New("invalid principal")

type Service struct {
	repo  *Repository
	now   func() time.Time
	newID func() string
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:  repo,
		now:   func() time.Time { return time.Now().UTC() },
		newID: uuid.NewString,
	}
}

func (s *Service) ResolveOrProvisionUser(ctx context.Context, principal ExternalPrincipal) (*ResolvedIdentity, error) {
	if err := validatePrincipal(principal); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetResolvedByIssuerSubject(ctx, normalizeIssuer(principal.Issuer), normalizeSubject(principal.Subject))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if err == nil && existing != nil {
		if err := s.refreshResolvedIdentity(ctx, existing, principal); err != nil {
			return nil, err
		}
		return s.repo.GetResolvedByIssuerSubject(ctx, normalizeIssuer(principal.Issuer), normalizeSubject(principal.Subject))
	}

	user := principalToUser(s.newID(), principal)
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	identity := principalToExternalIdentity(s.newID(), user.ID, principal)
	if err := s.repo.CreateExternalIdentity(ctx, identity); err != nil {
		return nil, err
	}

	return &ResolvedIdentity{
		User:             user,
		ExternalIdentity: identity,
	}, nil
}

func (s *Service) refreshResolvedIdentity(ctx context.Context, resolved *ResolvedIdentity, principal ExternalPrincipal) error {
	if resolved == nil || resolved.User == nil || resolved.ExternalIdentity == nil {
		return fmt.Errorf("resolved identity is incomplete")
	}

	resolved.User.PrimaryEmail = normalizeEmail(principal.Email)
	resolved.User.DisplayName = displayNameForPrincipal(principal)
	resolved.User.AvatarURL = strings.TrimSpace(principal.AvatarURL)
	if err := s.repo.UpdateUserProfile(ctx, resolved.User); err != nil {
		return err
	}

	claimsJSON, err := marshalClaims(principal.Claims)
	if err != nil {
		return err
	}

	resolved.ExternalIdentity.Email = normalizeEmail(principal.Email)
	resolved.ExternalIdentity.EmailVerified = principal.EmailVerified
	resolved.ExternalIdentity.PreferredUsername = strings.TrimSpace(principal.PreferredUsername)
	resolved.ExternalIdentity.RawClaimsJSON = claimsJSON
	if providerKind := normalizeProviderKind(principal.ProviderKind); providerKind != "" {
		resolved.ExternalIdentity.ProviderKind = providerKind
	}

	return s.repo.UpdateExternalIdentity(ctx, resolved.ExternalIdentity)
}

func validatePrincipal(principal ExternalPrincipal) error {
	if normalizeIssuer(principal.Issuer) == "" {
		return fmt.Errorf("%w: issuer is required", ErrInvalidPrincipal)
	}
	if normalizeSubject(principal.Subject) == "" {
		return fmt.Errorf("%w: subject is required", ErrInvalidPrincipal)
	}
	return nil
}

func principalToUser(id string, principal ExternalPrincipal) *User {
	return &User{
		ID:           id,
		PrimaryEmail: normalizeEmail(principal.Email),
		DisplayName:  displayNameForPrincipal(principal),
		AvatarURL:    strings.TrimSpace(principal.AvatarURL),
	}
}

func principalToExternalIdentity(id, userID string, principal ExternalPrincipal) *ExternalIdentity {
	claimsJSON, _ := marshalClaims(principal.Claims)

	return &ExternalIdentity{
		ID:                id,
		UserID:            userID,
		Issuer:            normalizeIssuer(principal.Issuer),
		Subject:           normalizeSubject(principal.Subject),
		ProviderKind:      normalizeProviderKind(principal.ProviderKind),
		Email:             normalizeEmail(principal.Email),
		EmailVerified:     principal.EmailVerified,
		PreferredUsername: strings.TrimSpace(principal.PreferredUsername),
		RawClaimsJSON:     claimsJSON,
	}
}

func normalizeIssuer(value string) string {
	return strings.TrimSpace(value)
}

func normalizeSubject(value string) string {
	return strings.TrimSpace(value)
}

func normalizeEmail(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func normalizeProviderKind(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ProviderKindOIDC
	}
	return value
}

func displayNameForPrincipal(principal ExternalPrincipal) string {
	if value := strings.TrimSpace(principal.DisplayName); value != "" {
		return value
	}
	if value := strings.TrimSpace(principal.PreferredUsername); value != "" {
		return value
	}
	if value := normalizeEmail(principal.Email); value != "" {
		return value
	}
	return normalizeSubject(principal.Subject)
}

func marshalClaims(claims map[string]any) (string, error) {
	if len(claims) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
