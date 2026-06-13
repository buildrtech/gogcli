package googleapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/steipete/gogcli/internal/config"
)

var errServiceAccountStoreRequired = errors.New("service account store resolver is required")

type serviceAccountStoreResolver func() (*config.ServiceAccountStore, error)

type serviceAccountStoreContextKey struct{}

func WithServiceAccountStoreResolver(ctx context.Context, resolver func() (*config.ServiceAccountStore, error)) context.Context {
	return context.WithValue(ctx, serviceAccountStoreContextKey{}, serviceAccountStoreResolver(resolver))
}

func serviceAccountStoreFromContext(ctx context.Context) (*config.ServiceAccountStore, error) {
	resolver, ok := ctx.Value(serviceAccountStoreContextKey{}).(serviceAccountStoreResolver)
	// Google auth composition must provide an explicit repository, even when
	// it is empty. Falling through here would hide incomplete runtime wiring.
	if !ok || resolver == nil {
		return nil, errServiceAccountStoreRequired
	}

	store, err := resolver()
	if err != nil {
		return nil, fmt.Errorf("resolve service account store: %w", err)
	}

	if store == nil {
		return nil, errServiceAccountStoreRequired
	}

	return store, nil
}

func serviceAccountSubject(subject string, serviceAccountEmail string) string {
	subject = strings.TrimSpace(subject)
	serviceAccountEmail = strings.TrimSpace(serviceAccountEmail)

	if subject == "" || strings.EqualFold(subject, serviceAccountEmail) {
		return ""
	}

	return subject
}

var newServiceAccountTokenSource = func(ctx context.Context, keyJSON []byte, subject string, scopes []string) (oauth2.TokenSource, error) {
	cfg, err := google.JWTConfigFromJSON(keyJSON, scopes...)
	if err != nil {
		return nil, fmt.Errorf("parse service account: %w", err)
	}
	// Only set Subject (impersonation) when the caller requests a different
	// identity than the service account itself. When subject matches the
	// SA's client_email we run in pure SA mode: no Domain-Wide Delegation.
	cfg.Subject = serviceAccountSubject(subject, cfg.Email)

	// Ensure token exchanges don't hang forever.
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: tokenExchangeTimeout})

	return cfg.TokenSource(ctx), nil
}

func tokenSourceForServiceAccountScopes(ctx context.Context, serviceLabel string, email string, scopes []string) (oauth2.TokenSource, string, bool, error) {
	store, err := serviceAccountStoreFromContext(ctx)
	if err != nil {
		return nil, "", false, err
	}

	file, exists, err := store.Read(email, serviceLabel == "keep")
	if err != nil {
		return nil, "", false, fmt.Errorf("read service account: %w", err)
	}

	if !exists {
		return nil, "", false, nil
	}

	ts, err := newServiceAccountTokenSource(ctx, file.Data, email, scopes)
	if err != nil {
		return nil, "", false, err
	}

	return ts, file.Path, true, nil
}
