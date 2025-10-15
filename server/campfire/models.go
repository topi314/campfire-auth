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

type Pagination[T any] struct {
	TotalCount int       `json:"totalCount"`
	Edges      []Edge[T] `json:"edges"`
	PageInfo   PageInfo  `json:"pageInfo"`
}

type Edge[T any] struct {
	Node   T      `json:"node"`
	Cursor string `json:"cursor"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	StartCursor string `json:"startCursor"`
	EndCursor   string `json:"endCursor"`
}
type clubResp struct {
	Club Club `json:"club"`
}

type Club struct {
	ID       string              `json:"id"`
	Channels Pagination[Channel] `json:"channels"`
}

type Channel struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	ChatV2TopicId       string `json:"chatV2TopicId"`
	UnreadMentionsCount int    `json:"unreadMentionsCount"`
	UserSettings        struct {
		IsMuted bool `json:"isMuted"`
	}
	IsReadyOnly bool `json:"isReadyOnly"`
}

type Member struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   string  `json:"avatarUrl"`
	Badges      []Badge `json:"badges"`
}

type Badge struct {
	Alias     string `json:"alias"`
	BadgeType string `json:"badgeType"`
}
