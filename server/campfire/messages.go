package campfire

import (
	"context"
	_ "embed"
)

//go:embed queries/message_history.graphql
var historyQuery string

func (c *Client) GetMessageHistory(ctx context.Context, channelID string) (*MessageHistory, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}

	var history historyResp
	if err = c.Do(ctx, token, historyQuery, map[string]any{
		"input": map[string]any{
			"channelId": channelID,
		},
	}, &history); err != nil {
		return nil, err
	}

	return &history.MessagesFromHistoryV2, nil
}
