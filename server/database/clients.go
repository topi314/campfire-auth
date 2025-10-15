package database

import (
	"context"
	"time"

	"github.com/topi314/campfire-auth/internal/xpgtype"
)

type Client struct {
	ID           string                 `db:"client_id"`
	Name         string                 `db:"client_name"`
	Secret       string                 `db:"client_secret"`
	RedirectURIs xpgtype.JSON[[]string] `db:"client_redirect_uris"`
	CreatedAt    time.Time              `db:"client_created_at"`
}

func (d *Database) InsertClient(ctx context.Context, name string, clientID string, clientSecret string, redirectURIs []string) error {
	query := `
		INSERT INTO clients (name, client_id, client_secret, client_redirect_uris)
		VALUES ($1, $2, $3, $4)
	`

	_, err := d.db.ExecContext(ctx, query, name, clientID, clientSecret, xpgtype.JSON[[]string]{V: redirectURIs})

	return err
}

func (d *Database) GetClient(ctx context.Context, clientID string) (*Client, error) {
	query := `
		SELECT *
		FROM clients
		WHERE client_id = $1
	`

	var client Client
	if err := d.db.GetContext(ctx, &client, query, clientID); err != nil {
		return nil, err
	}

	return &client, nil
}

func (d *Database) GetClients(ctx context.Context) ([]Client, error) {
	query := `
		SELECT *
		FROM clients
		ORDER BY client_created_at DESC
	`

	var clients []Client
	if err := d.db.SelectContext(ctx, &clients, query); err != nil {
		return nil, err
	}

	return clients, nil
}

func (d *Database) GetClientByIDSecret(ctx context.Context, clientID, clientSecret string) (*Client, error) {
	query := `
		SELECT *
		FROM clients
		WHERE client_id = $1 AND client_secret = $2
	`

	var client Client
	if err := d.db.GetContext(ctx, &client, query, clientID, clientSecret); err != nil {
		return nil, err
	}

	return &client, nil
}
