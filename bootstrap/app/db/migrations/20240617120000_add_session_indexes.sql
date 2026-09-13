-- +goose Up
create unique index if not exists idx_sessions_token on sessions(token);
create index if not exists idx_sessions_expires_at on sessions(expires_at);
create index if not exists idx_sessions_user_id on sessions(user_id);

-- +goose Down
drop index if exists idx_sessions_user_id;
drop index if exists idx_sessions_expires_at;
drop index if exists idx_sessions_token;
