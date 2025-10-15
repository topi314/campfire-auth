package database

import (
	"context"
	"time"

	"github.com/topi314/campfire-auth/internal/xpgtype"
)

type Client struct {
	ID           int                    `db:"client_id"`
	Secret       string                 `db:"client_secret"`
	ClubID       string                 `db:"client_club_id"`
	ChannelID    string                 `db:"client_channel_id"`
	RedirectURIs xpgtype.JSON[[]string] `db:"client_redirect_uris"`
	CreatedAt    time.Time              `db:"client_created_at"`
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
