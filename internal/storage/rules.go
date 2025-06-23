package storage

import (
	"context"

	"inbox451/internal/models"
)

func (r *repository) ListRules(ctx context.Context, limit, offset int) ([]*models.ForwardRule, int, error) {
	var total int
	err := r.queries.CountRules.GetContext(ctx, &total)
	if err != nil {
		return nil, 0, err
	}

	rules := []*models.ForwardRule{}
	if total > 0 {
		err = r.queries.ListRules.SelectContext(ctx, &rules, limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	return rules, total, nil
}

func (r *repository) ListRulesByInbox(ctx context.Context, inboxID string, limit, offset int) ([]*models.ForwardRule, int, error) {
	var total int
	err := r.queries.CountRulesByInbox.GetContext(ctx, &total, inboxID)
	if err != nil {
		return nil, 0, err
	}

	rules := []*models.ForwardRule{}
	if total > 0 {
		err = r.queries.ListRulesByInbox.SelectContext(ctx, &rules, inboxID, limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	return rules, total, nil
}

func (r *repository) GetRule(ctx context.Context, id string) (*models.ForwardRule, error) {
	var rule models.ForwardRule
	err := r.queries.GetRule.GetContext(ctx, &rule, id)
	return &rule, handleDBError(err)
}

func (r *repository) CreateRule(ctx context.Context, rule *models.ForwardRule) error {
	return r.queries.CreateRule.QueryRowContext(ctx, rule.InboxID, rule.Sender, rule.Receiver, rule.Subject).
		Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
}

func (r *repository) UpdateRule(ctx context.Context, rule *models.ForwardRule) error {
	result, err := r.queries.UpdateRule.ExecContext(ctx, rule.Sender, rule.Receiver, rule.Subject, rule.ID)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

func (r *repository) DeleteRule(ctx context.Context, id string) error {
	result, err := r.queries.DeleteRule.ExecContext(ctx, id)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}
