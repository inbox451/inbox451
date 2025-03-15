package core

import (
	"fmt"
	"os"

	"inbox451/internal/config"
	"inbox451/internal/logger"
	"inbox451/internal/models"
	"inbox451/internal/storage"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type Core struct {
	Config     *config.Config
	Logger     *logger.Logger
	Repository storage.Repository
	Version    string
	Commit     string
	BuildDate  string

	UserService    UserService
	TokenService   TokenService
	ProjectService ProjectService
	InboxService   InboxService
	RuleService    RuleService
	MessageService MessageService
}

func NewCore(cfg *config.Config, db *sqlx.DB, version, commit, date string) (*Core, error) {
	baseLogger := logger.New(os.Stdout, cfg.Logging.Level)

	repo, err := storage.NewRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	core := &Core{
		Config:     cfg,
		Logger:     baseLogger,
		Repository: repo,
		Version:    version,
		Commit:     commit,
		BuildDate:  date,
	}

	core.UserService = NewUserService(core)
	core.ProjectService = NewProjectService(core)
	core.InboxService = NewInboxService(core)
	core.RuleService = NewRuleService(core)
	core.MessageService = NewMessageService(core)
	core.TokenService = NewTokensService(core)

	return core, nil
}

func (c *Core) StoreMessage(message models.Message) (models.Message, error) {
	return c.MessageService.Store(message)
}
