package core

import (
	"context"

	"inbox451/internal/models"
)

type RuleService struct {
	core *Core
}

func NewRuleService(core *Core) RuleService {
	return RuleService{core: core}
}

func (s *RuleService) Create(ctx context.Context, rule *models.ForwardRule) error {
	s.core.Logger.Info("Creating new rule for inbox %d", rule.InboxID)

	if err := s.core.Repository.CreateRule(ctx, rule); err != nil {
		s.core.Logger.Error("Failed to create rule: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully created rule with ID: %d", rule.ID)
	return nil
}

func (s *RuleService) Get(ctx context.Context, id string) (*models.ForwardRule, error) {
	s.core.Logger.Debug("Fetching rule with ID: %d", id)

	rule, err := s.core.Repository.GetRule(ctx, id)
	if err != nil {
		s.core.Logger.Error("Failed to fetch rule: %v", err)
		return nil, err
	}

	if rule == nil {
		s.core.Logger.Info("Rule not found with ID: %d", id)
		return nil, ErrNotFound
	}

	return rule, nil
}

func (s *RuleService) Update(ctx context.Context, rule *models.ForwardRule) error {
	s.core.Logger.Info("Updating rule with ID: %d", rule.ID)

	if err := s.core.Repository.UpdateRule(ctx, rule); err != nil {
		s.core.Logger.Error("Failed to update rule: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully updated rule with ID: %d", rule.ID)
	return nil
}

func (s *RuleService) Delete(ctx context.Context, id string) error {
	s.core.Logger.Info("Deleting rule with ID: %d", id)

	if err := s.core.Repository.DeleteRule(ctx, id); err != nil {
		s.core.Logger.Error("Failed to delete rule: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted rule with ID: %d", id)
	return nil
}

func (s *RuleService) ListByInbox(ctx context.Context, inboxID string, limit, offset int) (*models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing rules for inbox %d with limit: %d and offset: %d", inboxID, limit, offset)

	rules, total, err := s.core.Repository.ListRulesByInbox(ctx, inboxID, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list rules: %v", err)
		return nil, err
	}

	response := &models.PaginatedResponse{
		Data: rules,
	}
	response.Pagination.Total = total
	response.Pagination.Limit = limit
	response.Pagination.Offset = offset

	s.core.Logger.Info("Successfully retrieved %d rules (total: %d)", len(rules), total)
	return response, nil
}
