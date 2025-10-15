package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/topi314/campfire-auth/server/campfire"
	"github.com/topi314/campfire-auth/server/database"
)

func (s *Server) loginCodeChecker() {
	for {
		s.doLoginCodeCheck()
		time.Sleep(1 * time.Second)
	}
}

func (s *Server) doLoginCodeCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logins, err := s.DB.GetNextLogins(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get next login", slog.String("err", err.Error()))
		return
	}
	logins = filterLogins(logins)
	if len(logins) == 0 {
		time.Sleep(1 * time.Second)
		return
	}

	if err = s.handleLoginCheck(ctx, logins); err != nil {
		ids := make([]int, 0, len(logins))
		for _, login := range logins {
			ids = append(ids, login.ID)
		}
		if err = s.DB.UpdateLoginsLastUpdatedAt(ctx, ids); err != nil {
			slog.ErrorContext(ctx, "Failed to update login last updated at", slog.String("err", err.Error()))
		}
		return
	}
}

func filterLogins(logins []database.Login) []database.Login {
	if len(logins) == 0 {
		return nil
	}

	first := logins[0]
	sameChannelLogins := make([]database.Login, 0, len(logins))
	for _, login := range logins {
		if login.ChannelID == first.ChannelID {
			sameChannelLogins = append(sameChannelLogins, login)
		}
	}

	return sameChannelLogins
}

func (s *Server) handleLoginCheck(ctx context.Context, logins []database.Login) error {
	members, err := s.checkForCode(ctx, logins)
	if err != nil {
		return err
	}

	updates := make(map[int]json.RawMessage, len(logins))
	for id, member := range members {
		memberData, err := json.Marshal(member)
		if err != nil {
			return err
		}

		updates[id] = memberData
	}

	return s.DB.UpdateLoginUsers(ctx, updates)
}

func (s *Server) checkForCode(ctx context.Context, logins []database.Login) (map[int]campfire.User, error) {
	client := logins[0]

	history, err := s.Campfire.GetMessageHistory(ctx, client.ChannelID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get message history", slog.String("err", err.Error()))
		return nil, err
	}

	users := make(map[int]campfire.User)
	for _, login := range logins {
		for _, message := range history.Messages {
			if strings.Contains(message.Message.Content, login.Code) {
				users[login.ID] = message.Message.Sender.User
				break
			}
		}
	}

	return users, nil
}
