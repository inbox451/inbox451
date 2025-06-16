--- ------------------------------------------
-- Users
-- -------------------------------------------

-- name: list-projects
SELECT id, name, created_at, updated_at
FROM projects
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: list-projects-by-user
SELECT projects.id, projects.name, projects.created_at, projects.updated_at
FROM projects
INNER JOIN project_users ON projects.id = project_users.project_id
WHERE project_users.user_id = $1
ORDER BY projects.id
LIMIT $2 OFFSET $3;

-- name: count-projects-by-user
SELECT COUNT(DISTINCT(projects.id))
FROM projects
INNER JOIN project_users ON projects.id = project_users.project_id
WHERE project_users.user_id = $1;

-- name: get-project
SELECT id, name, created_at, updated_at
FROM projects
WHERE id = $1;

-- name: create-project
INSERT INTO projects (name, created_at, updated_at)
VALUES ($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, created_at, updated_at;

-- name: update-project
UPDATE projects
SET name = $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2
RETURNING updated_at;

-- name: delete-project
DELETE FROM projects WHERE id = $1;

-- name: count-projects
SELECT COUNT(*) FROM projects;

--- ------------------------------------------
-- Project Users
-- -------------------------------------------

-- name: add-user-to-project
INSERT INTO project_users (user_id, project_id, role)
VALUES ($1, $2, $3)
RETURNING created_at, updated_at;

-- name: remove-user-from-project
DELETE FROM project_users
WHERE user_id = $1 AND project_id = $2;

--- ------------------------------------------
-- Inboxes
-- -------------------------------------------

-- name: create-inbox
INSERT INTO inboxes (project_id, email, created_at, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, created_at, updated_at;

-- name: get-inbox
SELECT id, project_id, email, created_at, updated_at
FROM inboxes
WHERE id = $1;

-- name: update-inbox
UPDATE inboxes
SET email = $1
WHERE id = $2;

-- name: delete-inbox
DELETE FROM inboxes WHERE id = $1;

-- name: list-inboxes-by-project
SELECT id, project_id, email, created_at, updated_at
FROM inboxes
WHERE project_id = $1
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: count-inboxes-by-project
SELECT COUNT(*)
FROM inboxes
WHERE project_id = $1;

-- name: get-inbox-by-email
SELECT id, project_id, email, created_at, updated_at
FROM inboxes
WHERE email = $1;

--- ------------------------------------------
-- Rules
-- -------------------------------------------

-- name: create-rule
INSERT INTO forward_rules (inbox_id, sender, receiver, subject, created_at, updated_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, created_at, updated_at;

-- name: get-rule
SELECT id, inbox_id, sender, receiver, subject, created_at, updated_at
FROM forward_rules
WHERE id = $1;

-- name: update-rule
UPDATE forward_rules
SET sender = $1, receiver = $2, subject = $3
WHERE id = $4;

-- name: delete-rule
DELETE FROM forward_rules WHERE id = $1;

-- name: list-rules-by-inbox
SELECT id, inbox_id, sender, receiver, subject, created_at, updated_at
FROM forward_rules
WHERE inbox_id = $1
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: count-rules-by-inbox
SELECT COUNT(*)
FROM forward_rules
WHERE inbox_id = $1;

-- name: list-rules
SELECT id, inbox_id, sender, receiver, subject, created_at, updated_at
FROM forward_rules
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: count-rules
SELECT COUNT(*) FROM forward_rules;

--- ------------------------------------------
-- Messages
-- -------------------------------------------

-- name: create-message
INSERT INTO messages (inbox_id, sender, receiver, subject, body, is_read, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, created_at, updated_at;

-- name: get-message
SELECT id, inbox_id, sender, receiver, subject, body, is_read, created_at, updated_at
FROM messages
WHERE id = $1;

-- name: list-messages-by-inbox
SELECT id, inbox_id, sender, receiver, subject, body, is_read, created_at, updated_at
FROM messages
WHERE inbox_id = $1
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: count-messages-by-inbox
SELECT COUNT(*)
FROM messages
WHERE inbox_id = $1;

-- name: update-message-read-status
UPDATE messages
SET is_read = $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2;

-- name: delete-message
DELETE FROM messages WHERE id = $1;

-- name: list-messages-by-inbox-with-read-filter
SELECT id, inbox_id, sender, receiver, subject, body, is_read, created_at, updated_at
FROM messages
WHERE inbox_id = $1 AND is_read = $2
ORDER BY id
LIMIT $3 OFFSET $4;

-- name: count-messages-by-inbox-with-read-filter
SELECT COUNT(*)
FROM messages
WHERE inbox_id = $1 AND is_read = $2;

--- ------------------------------------------
-- Users
-- -------------------------------------------

-- name: list-users
SELECT id, name, username, password, email, status, role,
       loggedin_at, created_at, updated_at
FROM users
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: count-users
SELECT COUNT(*)
FROM users

-- name: create-user
INSERT INTO users (name, username, password, email, status, role, password_login, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, created_at, updated_at;

-- name: get-user
SELECT id, name, username, password, email, status, role, password_login, loggedin_at, created_at, updated_at
FROM users
WHERE id = $1;

-- name: get-user-by-email
SELECT id, name, username, password, email, status, role, password_login, loggedin_at, created_at, updated_at
FROM users
WHERE email = $1;

-- name: update-user
UPDATE users
SET name = $1, username = $2, password = CASE WHEN $3 = '' THEN password ELSE $3 END, email = $4, status = $5, role = $6, password_login = $7, updated_at = NOW()
WHERE id = $8
RETURNING updated_at;

-- name: delete-user
DELETE FROM users WHERE id = $1;

-- name: get-user-by-username
SELECT id, name, username, password, email, status, role, password_login, loggedin_at, created_at, updated_at
FROM users
WHERE username = $1;

--- ------------------------------------------
-- Tokens
-- -------------------------------------------

-- name: list-tokens-by-user
SELECT id, user_id, token, name, expires_at, last_used_at, created_at, updated_at
FROM tokens
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: count-tokens-by-user
SELECT COUNT(*) FROM tokens WHERE user_id = $1;

-- name: get-token-by-user
SELECT id, user_id, token, name, expires_at, last_used_at, created_at, updated_at
FROM tokens
WHERE id = $1 AND user_id = $2;

-- name: get-token-by-value
SELECT id, user_id, token, name, expires_at, last_used_at, created_at, updated_at
FROM tokens
WHERE token = $1

-- name: create-token
INSERT INTO tokens (user_id, token, name, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token, name, expires_at, created_at, updated_at;

-- name: delete-token
DELETE FROM tokens WHERE id = $1;

-- name: update-token-last-used
UPDATE tokens SET last_used_at = NOW() WHERE id = $1;

-- name: prune-expired-tokens
DELETE FROM tokens WHERE expires_at IS NOT NULL AND expires_at < NOW();


-- =============================================================================
-- Session queries needed by github.com/zerodha/simplesessions/stores/postgres
-- =============================================================================

-- name: get-session
SELECT data FROM sessions WHERE id=$1;

-- name: insert-session
INSERT INTO sessions (id, data, created_at) VALUES ($1, $2, now());

-- name: delete-session
DELETE FROM sessions WHERE id=$1;

-- name: delete-expired-sessions
DELETE FROM sessions;

-- name: update-session
UPDATE sessions SET data=$1  WHERE id=$2;

-- =============================================================================
-- IMAP-related queries
-- =============================================================================

-- name: update-message-deleted-status
UPDATE messages
SET is_deleted = $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2;

-- name: list-messages-by-inbox-with-filters
SELECT id, inbox_id, sender, receiver, subject, body, is_read, is_deleted, created_at, updated_at
FROM messages
WHERE inbox_id = $1
  AND ($2::BOOLEAN IS NULL OR is_read = $2)
  AND ($3::BOOLEAN IS NULL OR is_deleted = $3)
ORDER BY id
LIMIT $4 OFFSET $5;

-- name: count-messages-by-inbox-with-filters
SELECT COUNT(*)
FROM messages
WHERE inbox_id = $1
  AND ($2::BOOLEAN IS NULL OR is_read = $2)
  AND ($3::BOOLEAN IS NULL OR is_deleted = $3);

-- name: list-inboxes-by-user
SELECT DISTINCT i.id, i.project_id, i.email, i.created_at, i.updated_at
FROM inboxes i
INNER JOIN project_users pu ON i.project_id = pu.project_id
WHERE pu.user_id = $1
ORDER BY i.email;

-- name: get-inbox-by-email-and-user
SELECT DISTINCT i.id, i.project_id, i.email, i.created_at, i.updated_at
FROM inboxes i
INNER JOIN project_users pu ON i.project_id = pu.project_id
WHERE i.email = $1 AND pu.user_id = $2;

-- name: get-messages-by-uids
SELECT id, inbox_id, sender, receiver, subject, body, is_read, is_deleted, created_at, updated_at
FROM messages
WHERE inbox_id = $1 AND id = ANY($2::int[])
ORDER BY id;

-- name: get-all-message-uids-for-inbox
SELECT id
FROM messages
WHERE inbox_id = $1 AND is_deleted = false
ORDER BY id;

-- name: get-all-message-uids-for-inbox-including-deleted
SELECT id
FROM messages
WHERE inbox_id = $1
ORDER BY id;

-- name: get-max-message-uid
SELECT COALESCE(MAX(id), 0) AS max_uid
FROM messages
WHERE inbox_id = $1;
