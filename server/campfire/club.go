package campfire

import (
	"encoding/base64"
	"fmt"
	"net/url"
)

func ResolveClubAndChannelID(clubURL string) (string, string, error) {
	u, err := url.Parse(clubURL)
	if err != nil {
		return "", "", err
	}

	query := u.Query()
	sub := query.Get("deep_link_sub1")
	if sub == "" {
		return "", "", fmt.Errorf("no 'deep_link_sub1' query parameter found in URL: %s", clubURL)
	}

	decoded, err := base64.StdEncoding.DecodeString(sub)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode base64 string: %w", err)
	}

	values, err := url.ParseQuery(string(decoded))
	if err != nil {
		return "", "", fmt.Errorf("failed to parse decoded string as query parameters: %w", err)
	}

	r := values.Get("r")
	if r != "clubs" {
		return "", "", fmt.Errorf("unexpected 'r' parameter value in decoded param: %s", string(decoded))
	}

	clubID := values.Get("c")
	if clubID == "" {
		return "", "", fmt.Errorf("no 'c' parameter found in decoded param: %s", string(decoded))
	}
	channelID := values.Get("ch")
	if channelID == "" {
		return "", "", fmt.Errorf("no 'ch' parameter found in decoded param: %s", string(decoded))
	}

	return clubID, channelID, nil
}
