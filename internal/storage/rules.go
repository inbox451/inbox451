package storage

import (
	"inbox451/internal/models"
)

func (r *repository) ListRules(limit, offset int) ([]models.ForwardRule, int, error) {
	rules := []models.ForwardRule{}
	var total int

	err := r.queries.CountRules.Get(&total)
	if err != nil {
		return nil, 0, err
	}

	err = r.queries.ListRules.Select(&rules, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

func (r *repository) ListRulesByInbox(inboxId, limit, offset int) ([]models.ForwardRule, int, error) {
	rules := []models.ForwardRule{}
	var total int

	err := r.queries.CountRulesByInbox.Get(&total, inboxId)
	if err != nil {
		return nil, 0, err
	}

	err = r.queries.ListRulesByInbox.Select(&rules, inboxId, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return rules, total, nil
}

func (r *repository) GetRule(id int) (models.ForwardRule, error) {
	var rule models.ForwardRule
	err := r.queries.GetRule.Get(&rule, id)
	return rule, handleDBError(err)
}

func (r *repository) CreateRule(rule models.ForwardRule) (models.ForwardRule, error) {
	var ruleId int

	err := r.queries.CreateRule.QueryRow(
		rule.InboxID,
		rule.Sender,
		rule.Receiver,
		rule.Subject).Scan(&ruleId)

	if err != nil {
		return models.ForwardRule{}, handleDBError(err)
	}

	return r.GetRule(ruleId)
}

func (r *repository) UpdateRule(rule models.ForwardRule) (models.ForwardRule, error) {
	result, err := r.queries.UpdateRule.Exec(rule.Sender, rule.Receiver, rule.Subject, rule.ID)

	if err != nil {
		return models.ForwardRule{}, handleDBError(err)
	}

	if err := handleRowsAffected(result); err != nil {
		return models.ForwardRule{}, err
	}

	return r.GetRule(rule.ID)
}

func (r *repository) DeleteRule(ruleId int) error {
	result, err := r.queries.DeleteRule.Exec(ruleId)

	if err != nil {
		return handleDBError(err)
	}

	return handleRowsAffected(result)
}
