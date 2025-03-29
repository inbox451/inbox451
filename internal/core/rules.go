package core

import (
	"inbox451/internal/models"
)

type RuleService struct {
	core *Core
}

func NewRuleService(core *Core) RuleService {
	return RuleService{core: core}
}

func (s *RuleService) Create(rule models.ForwardRule) (models.ForwardRule, error) {
	s.core.Logger.Info("Creating new rule for inbox %d", rule.InboxID)

	rule, err := s.core.Repository.CreateRule(rule)
	if err != nil {
		s.core.Logger.Error("Failed to create rule: %v", err)
		return rule, err
	}

	s.core.Logger.Info("Successfully created rule with ID: %d", rule.ID)
	return rule, nil
}

func (s *RuleService) Get(id int) (models.ForwardRule, error) {
	s.core.Logger.Debug("Fetching rule with ID: %d", id)

	rule, err := s.core.Repository.GetRule(id)
	if err != nil {
		s.core.Logger.Error("Failed to fetch rule: %v", err)
		return rule, err
	}

	return rule, nil
}

func (s *RuleService) Update(rule models.ForwardRule) (models.ForwardRule, error) {
	s.core.Logger.Info("Updating rule with ID: %d", rule.ID)

	rule, err := s.core.Repository.UpdateRule(rule)
	if err != nil {
		s.core.Logger.Error("Failed to update rule: %v", err)
		return models.ForwardRule{}, err
	}

	s.core.Logger.Info("Successfully updated rule with ID: %d", rule.ID)
	return rule, nil
}

func (s *RuleService) Delete(id int) error {
	s.core.Logger.Info("Deleting rule with ID: %d", id)

	if err := s.core.Repository.DeleteRule(id); err != nil {
		s.core.Logger.Error("Failed to delete rule: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted rule with ID: %d", id)
	return nil
}

func (s *RuleService) ListByInbox(inboxID, limit, offset int) (models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing rules for inbox %d with limit: %d and offset: %d", inboxID, limit, offset)

	rules, total, err := s.core.Repository.ListRulesByInbox(inboxID, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list rules: %v", err)
		return models.PaginatedResponse{}, err
	}

	response := models.PaginatedResponse{
		Data: rules,
	}
	response.Pagination.Total = total
	response.Pagination.Limit = limit
	response.Pagination.Offset = offset

	s.core.Logger.Info("Successfully retrieved %d rules (total: %d)", len(rules), total)
	return response, nil
}
