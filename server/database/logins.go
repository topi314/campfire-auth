package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Login struct {
	ID           int              `db:"login_id"`
	ClientID     string           `db:"login_client_id"`
	Code         string           `db:"login_code"`
	CheckCode    string           `db:"login_check_code"`
	ExchangeCode string           `db:"login_exchange_code"`
	RedirectURI  string           `db:"login_redirect_uri"`
	State        string           `db:"login_state"`
	User         *json.RawMessage `db:"login_user"`
	CreatedAt    time.Time        `db:"login_created_at"`
	UpdatedAt    time.Time        `db:"login_updated_at"`
}

type LoginWithClient struct {
	Login
	Client
}

func (d *Database) InsertLogin(ctx context.Context, clientID, code string, checkCode string, exchangeCode string, redirectURI string, state string) error {
	query := `
		INSERT INTO logins (login_client_id, login_code, login_check_code, login_exchange_code, login_redirect_uri, login_state)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := d.db.ExecContext(ctx, query, clientID, code, checkCode, exchangeCode, redirectURI, state); err != nil {
		return fmt.Errorf("failed to insert login: %w", err)
	}

	return nil
}

func (d *Database) GetLoginByCheckCode(ctx context.Context, checkCode string) (*LoginWithClient, error) {
	query := `
		SELECT logins.*, clients.*
		FROM logins
		JOIN clients ON logins.login_client_id = clients.client_id
		WHERE logins.login_check_code = $1
	`

	var login LoginWithClient
	if err := d.db.GetContext(ctx, &login, query, checkCode); err != nil {
		return nil, fmt.Errorf("failed to get login by check code: %w", err)
	}

	return &login, nil
}

func (d *Database) GetLoginByCode(ctx context.Context, code string) (*LoginWithClient, error) {
	query := `
		SELECT logins.*, clients.*
		FROM logins
		JOIN clients ON logins.login_client_id = clients.client_id
		WHERE logins.login_code = $1
	`

	var login LoginWithClient
	if err := d.db.GetContext(ctx, &login, query, code); err != nil {
		return nil, fmt.Errorf("failed to get login by code: %w", err)
	}

	return &login, nil
}

func (d *Database) UpdateLoginUsers(ctx context.Context, logins map[int]json.RawMessage) error {
	for id, user := range logins {
		if err := d.UpdateLoginUser(ctx, id, user); err != nil {
			return fmt.Errorf("failed to update login user for id %d: %w", id, err)
		}
	}
	return nil
}

func (d *Database) UpdateLoginUser(ctx context.Context, id int, user json.RawMessage) error {
	query := `
		UPDATE logins
		SET login_user = $2
		WHERE login_id = $1
	`

	if _, err := d.db.ExecContext(ctx, query, id, user); err != nil {
		return fmt.Errorf("failed to update login user: %w", err)
	}

	return nil
}

func (d *Database) DeleteLoginByClientIDSecretExchangeCode(ctx context.Context, clientID, clientSecret, exchangeCode string) (*Login, error) {
	query := `
		DELETE FROM logins
		USING clients
		WHERE logins.login_client_id = clients.client_id
		AND clients.client_id = $1
		AND clients.client_secret = $2
		AND logins.login_exchange_code = $3
		RETURNING logins.*
	`

	var login Login
	if err := d.db.GetContext(ctx, &login, query, clientID, clientSecret, exchangeCode); err != nil {
		return nil, fmt.Errorf("failed to delete login by client ID, secret and exchange code: %w", err)
	}

	return &login, nil
}

// GetNextLogins retrieves all logins from the next client that needs to be checked.
func (d *Database) GetNextLogins(ctx context.Context) ([]LoginWithClient, error) {
	query := `
		SELECT logins.*, clients.*
		FROM logins
		JOIN clients ON logins.login_client_id = clients.client_id
		WHERE logins.login_user IS NULL
		ORDER BY logins.login_updated_at ASC
	`

	var logins []LoginWithClient
	if err := d.db.SelectContext(ctx, &logins, query); err != nil {
		return nil, fmt.Errorf("failed to get next logins: %w", err)
	}

	return logins, nil
}

func (d *Database) UpdateLoginsLastUpdatedAt(ctx context.Context, ids []int) error {
	query := `
		UPDATE logins
		SET login_updated_at = now()
		WHERE login_id = ANY($1)
	`

	if _, err := d.db.ExecContext(ctx, query, ids); err != nil {
		return fmt.Errorf("failed to update logins last updated at: %w", err)
	}

	return nil
}

func (d *Database) DeleteExpiredLogins(ctx context.Context) error {
	query := `
		DELETE FROM logins
		WHERE login_created_at < now() - INTERVAL '120 seconds'
	`

	if _, err := d.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to delete expired logins: %w", err)
	}

	return nil
}
