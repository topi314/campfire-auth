package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/topi314/campfire-auth/server/campfire"
	"github.com/topi314/campfire-auth/server/database"
)

func (s *Server) check() {
	for {
		s.doCheck()
		time.Sleep(1 * time.Second)
	}
}

func (s *Server) doCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	login, err := s.DB.GetNextLogin(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		slog.ErrorContext(ctx, "Failed to get next login", slog.String("err", err.Error()))
		return
	}

	if err = s.handleLogin(ctx, *login); err != nil {
		if err = s.DB.UpdateLoginLastUpdatedAt(ctx, login.Login.ID); err != nil {
			slog.ErrorContext(ctx, "Failed to update login last updated at", slog.String("err", err.Error()))
		}
		return
	}
}

func (s *Server) handleLogin(ctx context.Context, login database.LoginWithClient) error {
	member, err := checkForCode(login)
	if err != nil {
		return err
	}

	memberData, err := json.Marshal(member)
	if err != nil {
		return err
	}

	return s.DB.UpdateLoginUser(ctx, login.Login.ID, memberData)
}

func checkForCode(login database.LoginWithClient) (campfire.Member, error) {
	time.Sleep(3 * time.Second)
	return campfire.Member{
		ID:          "E:3I7ZXKS4BN252MFQQ6GX7ROZOPJNA3RITFEIPUZGJ324ESDJ2RVA",
		Username:    "topi314",
		DisplayName: "topi",
		AvatarURL:   "",
		Badges:      nil,
	}, nil
}
