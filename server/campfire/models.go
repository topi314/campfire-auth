package campfire

import (
	"fmt"
	"strings"
)

type Req struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type Resp[T any] struct {
	Errors []Error `json:"errors"`
	Data   T       `json:"data"`
}

type Error struct {
	Message string `json:"message"`
	Path    []any  `json:"path"`
}

func (e Error) String() string {
	return e.Error()
}

func (e Error) Error() string {
	msg := fmt.Sprintf("Error: %s", e.Message)
	if len(e.Path) > 0 {
		var path []string
		for _, p := range e.Path {
			path = append(path, fmt.Sprint(p))
		}
		msg += fmt.Sprintf(", Path: %v", strings.Join(path, "."))
	}
	return msg
}

type historyResp struct {
	MessagesFromHistoryV2 MessageHistory `json:"messagesFromHistoryV2"`
}

type MessageHistory struct {
	Messages []struct {
		Message struct {
			Id     string `json:"id"`
			Sender struct {
				User User `json:"user"`
			} `json:"sender"`
			SentAt  string `json:"sentAt"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"messages"`
}

type User struct {
	ID           string        `json:"id"`
	Username     string        `json:"username"`
	DisplayName  string        `json:"displayName"`
	AvatarURL    string        `json:"avatarUrl"`
	Badges       []Badge       `json:"badges"`
	GameProfiles []GameProfile `json:"gameProfiles"`
}

type Badge struct {
	Alias     string `json:"alias"`
	BadgeType string `json:"badgeType"`
}

type GameProfile struct {
	ID                    string `json:"id"`
	Game                  string `json:"game"`
	Codename              string `json:"codename"`
	DisplayName           string `json:"displayName"`
	Level                 int    `json:"level"`
	Faction               string `json:"faction"`
	FactionColor          string `json:"factionColor"`
	Visibility            string `json:"visibility"`
	LastPlayedTimestampMs int64  `json:"lastPlayedTimestampMs"`
}

type usersResp struct {
	Users []User `json:"users"`
}

type userByIDResp struct {
	User User `json:"userById"`
}
