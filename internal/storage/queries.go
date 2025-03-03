package storage

import (
	"embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/goyesql/v2"
	goyesqlx "github.com/knadh/goyesql/v2/sqlx"
)

//go:embed queries.sql
var queriesFS embed.FS

type Queries struct {
	// Project queries
	ListProjects        *sqlx.Stmt `query:"list-projects"`
	CountProjects       *sqlx.Stmt `query:"count-projects"`
	ListProjectsByUser  *sqlx.Stmt `query:"list-projects-by-user"`
	CountProjectsByUser *sqlx.Stmt `query:"count-projects-by-user"`
	GetProject          *sqlx.Stmt `query:"get-project"`
	CreateProject       *sqlx.Stmt `query:"create-project"`
	UpdateProject       *sqlx.Stmt `query:"update-project"`
	DeleteProject       *sqlx.Stmt `query:"delete-project"`

	// ProjectUser queries
	GetProjectUser        *sqlx.Stmt `query:"get-project-user"`
	AddUserToProject      *sqlx.Stmt `query:"add-user-to-project"`
	RemoveUserFromProject *sqlx.Stmt `query:"remove-user-from-project"`

	// Inbox queries
	CreateInbox           *sqlx.Stmt `query:"create-inbox"`
	GetInbox              *sqlx.Stmt `query:"get-inbox"`
	UpdateInbox           *sqlx.Stmt `query:"update-inbox"`
	DeleteInbox           *sqlx.Stmt `query:"delete-inbox"`
	ListInboxesByProject  *sqlx.Stmt `query:"list-inboxes-by-project"`
	CountInboxesByProject *sqlx.Stmt `query:"count-inboxes-by-project"`
	GetInboxByEmail       *sqlx.Stmt `query:"get-inbox-by-email"`

	// Rule queries
	CreateRule        *sqlx.Stmt `query:"create-rule"`
	GetRule           *sqlx.Stmt `query:"get-rule"`
	UpdateRule        *sqlx.Stmt `query:"update-rule"`
	DeleteRule        *sqlx.Stmt `query:"delete-rule"`
	ListRulesByInbox  *sqlx.Stmt `query:"list-rules-by-inbox"`
	CountRulesByInbox *sqlx.Stmt `query:"count-rules-by-inbox"`
	ListRules         *sqlx.Stmt `query:"list-rules"`
	CountRules        *sqlx.Stmt `query:"count-rules"`

	// Message queries
	CreateMessage                      *sqlx.Stmt `query:"create-message"`
	GetMessage                         *sqlx.Stmt `query:"get-message"`
	ListMessagesByInbox                *sqlx.Stmt `query:"list-messages-by-inbox"`
	CountMessagesByInbox               *sqlx.Stmt `query:"count-messages-by-inbox"`
	UpdateMessageReadStatus            *sqlx.Stmt `query:"update-message-read-status"`
	DeleteMessage                      *sqlx.Stmt `query:"delete-message"`
	ListMessagesByInboxWithReadFilter  *sqlx.Stmt `query:"list-messages-by-inbox-with-read-filter"`
	CountMessagesByInboxWithReadFilter *sqlx.Stmt `query:"count-messages-by-inbox-with-read-filter"`

	// User queries
	ListUsers         *sqlx.Stmt `query:"list-users"`
	CountUsers        *sqlx.Stmt `query:"count-users"`
	GetUser           *sqlx.Stmt `query:"get-user"`
	CreateUser        *sqlx.Stmt `query:"create-user"`
	UpdateUser        *sqlx.Stmt `query:"update-user"`
	DeleteUser        *sqlx.Stmt `query:"delete-user"`
	GetUserByUsername *sqlx.Stmt `query:"get-user-by-username"`

	// Tokens
	ListTokensByUser  *sqlx.Stmt `query:"list-tokens-by-user"`
	CountTokensByUser *sqlx.Stmt `query:"count-tokens-by-user"`
	GetTokenByUser    *sqlx.Stmt `query:"get-token-by-user"`
	DeleteToken       *sqlx.Stmt `query:"delete-token"`
	CreateToken       *sqlx.Stmt `query:"create-token"`
}

func PrepareQueries(db *sqlx.DB) (*Queries, error) {
	// Read queries from embedded file
	queryBytes, err := queriesFS.ReadFile("queries.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to read queries file: %w", err)
	}

	// Parse queries
	queries, err := goyesql.ParseBytes(queryBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse queries: %w", err)
	}

	// Prepare statements
	var q Queries
	if err := goyesqlx.ScanToStruct(&q, queries, db); err != nil {
		return nil, fmt.Errorf("failed to prepare queries: %w", err)
	}

	return &q, nil
}
