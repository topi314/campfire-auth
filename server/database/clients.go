package database

import (
	"context"
	"time"

	"github.com/topi314/campfire-auth/internal/xpgtype"
)

type Client struct {
	ID           string                 `db:"client_id"`
	Secret       string                 `db:"client_secret"`
	ClubID       string                 `db:"client_club_id"`
	ChannelID    string                 `db:"client_channel_id"`
	RedirectURIs xpgtype.JSON[[]string] `db:"client_redirect_uris"`
	CreatedAt    time.Time              `db:"client_created_at"`
}

func (d *Database) InsertClient(ctx context.Context, clientID string, clientSecret string, clubID string, channelID string, redirectURIs []string) error {
	query := `
		INSERT INTO clients (client_id, client_secret, client_club_id, client_channel_id, client_redirect_uris)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := d.db.ExecContext(ctx, query, clientID, clientSecret, clubID, channelID, xpgtype.JSON[[]string]{V: redirectURIs})
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
