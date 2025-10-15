CREATE TABLE campfire_tokens
(
    campfire_token_id         BIGSERIAL PRIMARY KEY,
    campfire_token_token      VARCHAR   NOT NULL,
    campfire_token_expires_at TIMESTAMP NOT NULL,
    campfire_token_email      VARCHAR   NOT NULL
);

CREATE TABLE clients
(
    client_id            VARCHAR PRIMARY KEY,
    client_secret        VARCHAR   NOT NULL,
    client_club_id       VARCHAR   NOT NULL,
    client_channel_id    VARCHAR   NOT NULL,
    client_redirect_uris JSONB     NOT NULL,
    client_created_at    TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE logins
(
    login_id            BIGSERIAL PRIMARY KEY,
    login_client_id     VARCHAR   NOT NULL REFERENCES clients (client_id) ON DELETE CASCADE,
    login_code          VARCHAR   NOT NULL UNIQUE,
    login_check_code    VARCHAR   NOT NULL UNIQUE,
    login_exchange_code VARCHAR   NOT NULL UNIQUE,
    login_redirect_uri  VARCHAR   NOT NULL,
    login_user          JSONB,
    login_created_at    TIMESTAMP NOT NULL DEFAULT now(),
    login_updated_at    TIMESTAMP NOT NULL DEFAULT now()
);