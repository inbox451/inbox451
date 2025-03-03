package storage

import (
	"fmt"

	"inbox451/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository interface {
	// Project operations
	ListProjects(limit, offset int) ([]models.Project, int, error)
	ListProjectsByUser(userID int, limit, offset int) ([]models.Project, int, error)
	GetProject(id int) (models.Project, error)
	CreateProject(project models.Project) (models.Project, error)
	UpdateProject(project models.Project) (models.Project, error)
	DeleteProject(id int) error

	// ProjectUser operations
	// This is a many-to-many relationship between projects and users
	ProjectAddUser(projectUser models.ProjectUser) (models.ProjectUser, error)
	ProjectRemoveUser(projectID int, userID int) error
	GetProjectUser(projectId int, userId int) (models.ProjectUser, error)

	// Inbox operations
	ListInboxesByProject(projectID, limit, offset int) ([]models.Inbox, int, error)
	GetInbox(id int) (models.Inbox, error)
	CreateInbox(inbox models.Inbox) (models.Inbox, error)
	UpdateInbox(inbox models.Inbox) (models.Inbox, error)
	DeleteInbox(id int) error

	// Rule operations
	ListRulesByInbox(inboxID, limit, offset int) ([]models.ForwardRule, int, error)
	GetRule(id int) (models.ForwardRule, error)
	CreateRule(rule models.ForwardRule) (models.ForwardRule, error)
	UpdateRule(rule models.ForwardRule) (models.ForwardRule, error)
	DeleteRule(id int) error

	// Message operations
	ListRules(limit, offset int) ([]models.ForwardRule, int, error)
	GetInboxByEmail(email string) (models.Inbox, error)
	GetMessage(id int) (models.Message, error)
	ListMessagesByInbox(inboxID, limit, offset int) ([]models.Message, int, error)
	ListMessagesByInboxWithFilter(inboxID int, isRead *bool, limit, offset int) ([]models.Message, int, error)
	CreateMessage(message models.Message) (models.Message, error)
	UpdateMessageReadStatus(messageID int, isRead bool) (models.Message, error)
	DeleteMessage(messageID int) error

	// User operations
	ListUsers(limit, offset int) ([]models.User, int, error)
	GetUser(id int) (models.User, error)
	GetUserByUsername(username string) (models.User, error)
	CreateUser(user models.User) (models.User, error)
	UpdateUser(user models.User) (models.User, error)
	DeleteUser(userId int) error

	// Tokens
	ListTokensByUser(userID int, limit, offset int) ([]models.Token, int, error)
	GetTokenByUser(userID int, tokenID int) (models.Token, error)
	CreateToken(token models.Token) (models.Token, error)
	DeleteToken(tokenID int) error
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
