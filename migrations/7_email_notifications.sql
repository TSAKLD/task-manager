-- +goose Up
CREATE TABLE email_notifications(
    email TEXT NOT NULL PRIMARY KEY REFERENCES users(email) ON DELETE CASCADE,
    subject text NOT NULL,
    created_at timestamptz not NULL
);

ALTER TABLE users ADD COLUMN vip_status TEXT NOT NULL DEFAULT 'inactive';

-- +goose Down
ALTER TABLE users DROP COLUMN vip_status;
DROP TABLE email_notifications;