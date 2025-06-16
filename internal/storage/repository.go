package storage

import (
	"context"
	"fmt"

	"inbox451/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository interface {
	// Project operations
	ListProjects(ctx context.Context, limit, offset int) ([]*models.Project, int, error)
	ListProjectsByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Project, int, error)
	GetProject(ctx context.Context, id string) (*models.Project, error)
	CreateProject(ctx context.Context, project *models.Project) error
	UpdateProject(ctx context.Context, project *models.Project) error
	DeleteProject(ctx context.Context, id string) error

	// ProjectUser operations
	// This is a many-to-many relationship between projects and users
	ProjectAddUser(ctx context.Context, projectUser *models.ProjectUser) error
	ProjectRemoveUser(ctx context.Context, projectID string, userID string) error

	// Inbox operations
	ListInboxesByProject(ctx context.Context, projectID string, limit, offset int) ([]*models.Inbox, int, error)
	GetInbox(ctx context.Context, id string) (*models.Inbox, error)
	GetInboxByEmail(ctx context.Context, email string) (*models.Inbox, error)
	CreateInbox(ctx context.Context, inbox *models.Inbox) error
	UpdateInbox(ctx context.Context, inbox *models.Inbox) error
	DeleteInbox(ctx context.Context, id string) error

	// Rule operations
	ListRulesByInbox(ctx context.Context, inboxID string, limit, offset int) ([]*models.ForwardRule, int, error)
	GetRule(ctx context.Context, id string) (*models.ForwardRule, error)
	CreateRule(ctx context.Context, rule *models.ForwardRule) error
	UpdateRule(ctx context.Context, rule *models.ForwardRule) error
	DeleteRule(ctx context.Context, id string) error

	// Message operations
	ListRules(ctx context.Context, limit, offset int) ([]*models.ForwardRule, int, error)
	GetMessage(ctx context.Context, id string) (*models.Message, error)
	ListMessagesByInbox(ctx context.Context, inboxID string, limit, offset int) ([]*models.Message, int, error)
	ListMessagesByInboxWithFilter(ctx context.Context, inboxID string, isRead *bool, limit, offset int) ([]*models.Message, int, error)
	CreateMessage(ctx context.Context, message *models.Message) error
	UpdateMessageReadStatus(ctx context.Context, messageID string, isRead bool) error
	DeleteMessage(ctx context.Context, messageID string) error

	// IMAP-related operations
	UpdateMessageDeletedStatus(ctx context.Context, messageID int, isDeleted bool) error
	ListMessagesByInboxWithFilters(ctx context.Context, inboxID int, filters models.MessageFilters, limit, offset int) ([]*models.Message, int, error)
	ListInboxesByUser(ctx context.Context, userID int) ([]*models.Inbox, error)
	GetInboxByEmailAndUser(ctx context.Context, email string, userID int) (*models.Inbox, error)
	GetMessagesByUIDs(ctx context.Context, inboxID int, uids []uint32) ([]*models.Message, error)
	GetAllMessageUIDsForInbox(ctx context.Context, inboxID int) ([]uint32, error)
	GetAllMessageUIDsForInboxIncludingDeleted(ctx context.Context, inboxID int) ([]uint32, error)
	GetMaxMessageUID(ctx context.Context, inboxID int) (uint32, error)

	// User operations
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int, error)
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, userId string) error

	// Tokens
	ListTokensByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Token, int, error)
	GetTokenByUser(ctx context.Context, userID string, tokenID string) (*models.Token, error)
	CreateToken(ctx context.Context, token *models.Token) error
	DeleteToken(ctx context.Context, tokenID string) error
	GetTokenByValue(ctx context.Context, tokenValue string) (*models.Token, error)
	UpdateTokenLastUsed(ctx context.Context, tokenID string) error
	PruneExpiredTokens(ctx context.Context) (int64, error)
}

type repository struct {
	db      *sqlx.DB
	queries *Queries
}

func NewRepository(db *sqlx.DB) (Repository, error) {
	// Then prepare queries
	queries, err := PrepareQueries(db)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare queries: %w", err)
	}

	return &repository{
		db:      db,
		queries: queries,
	}, nil
}
