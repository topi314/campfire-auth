ALTER TABLE clients
    DROP COLUMN client_club_id,
    DROP COLUMN client_channel_id,
    ADD COLUMN client_name VARCHAR NOT NULL DEFAULT 'Unnamed Client';

ALTER TABLE logins
    ADD COLUMN login_club_id    VARCHAR NOT NULL DEFAULT '',
    ADD COLUMN login_channel_id VARCHAR NOT NULL DEFAULT '';