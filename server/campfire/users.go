package campfire

import (
	"context"
	_ "embed"
)

var (
	//go:embed queries/user_by_id.graphql
	userByIDQuery string
	//go:embed queries/users.graphql
	usersQuery string
)

func (c *Client) GetUserByID(ctx context.Context, id string) (*User, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}

	var user userByIDResp
	if err = c.Do(ctx, token, userByIDQuery, map[string]any{
		"id": id,
	}, &user); err != nil {
		return nil, err
	}

	return &user.User, nil
}

func (c *Client) SearchUsers(ctx context.Context, username string) ([]User, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}

	var users usersResp
	if err = c.Do(ctx, token, usersQuery, map[string]any{
		"username": username,
	}, &users); err != nil {
		return nil, err
	}

	return users.Users, nil
}
