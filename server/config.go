package server

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/topi314/campfire-auth/internal/xtime"
	"github.com/topi314/campfire-auth/server/campfire"
	"github.com/topi314/campfire-auth/server/database"
)

func LoadConfig(cfgPath string) (Config, error) {
	file, err := os.Open(cfgPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	cfg := defaultConfig()
	if _, err = toml.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file: %w", err)
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Log: LogConfig{
			Level:     slog.LevelInfo,
			Format:    LogFormatText,
			AddSource: false,
		},
		Server: ServerConfig{
			Addr: ":8086",
		},
		Campfire: campfire.Config{
			Every:      xtime.Duration(1 * time.Second),
			Burst:      40,
			MaxRetries: 3,
		},
	}
}

type Config struct {
	Dev           bool                `toml:"dev"`
	Log           LogConfig           `toml:"log"`
	Server        ServerConfig        `toml:"server"`
	Database      database.Config     `toml:"database"`
	Campfire      campfire.Config     `toml:"campfire"`
	Notifications NotificationsConfig `toml:"notifications"`
}

func (c Config) String() string {
	return fmt.Sprintf("Dev: %t\nLog: %s\nServer: %s\nDatabase: %s\nCampfire: %s\nNotifications: %s",
		c.Dev,
		c.Log,
		c.Server,
		c.Database,
		c.Campfire,
		c.Notifications,
	)
}

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    LogFormat  `toml:"format"`
	AddSource bool       `toml:"add_source"`
}

func (c LogConfig) String() string {
	return fmt.Sprintf("\n Level: %s\n Format: %s\n AddSource: %t",
		c.Level,
		c.Format,
		c.AddSource,
	)
}

type ServerConfig struct {
	Addr          string `toml:"addr"`
	AdminPassword string `toml:"admin_password"`
	PublicURL     string `toml:"public_url"`
}

func (c ServerConfig) String() string {
	return fmt.Sprintf("\n Address: %s\n AdminPassword: %s\n PublicURL: %s",
		c.Addr,
		strings.Repeat("*", len(c.AdminPassword)),
		c.PublicURL,
	)
}

type NotificationsConfig struct {
	Enabled    bool   `toml:"enabled"`
	WebhookURL string `toml:"webhook_url"`
}

func (c NotificationsConfig) String() string {
	return fmt.Sprintf("\n Enabled: %t\n WebhookURL: %s",
		c.Enabled,
		c.WebhookURL,
	)
}
