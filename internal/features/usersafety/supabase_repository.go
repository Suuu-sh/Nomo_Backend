package usersafety

import (
	"context"
	"net/url"

	"github.com/yota/nomo/backend/internal/supabase"
)

type SupabaseRepository struct {
	client *supabase.Client
}

func NewSupabaseRepository(client *supabase.Client) *SupabaseRepository {
	return &SupabaseRepository{client: client}
}

func (r *SupabaseRepository) BlockUser(ctx context.Context, authToken string, relation UserRelation) (map[string]any, error) {
	payload := map[string]any{"blocker_user_id": relation.ActorUserID, "blocked_user_id": relation.TargetUserID}
	q := url.Values{}
	q.Set("on_conflict", "blocker_user_id,blocked_user_id")
	var rows []map[string]any
	if err := r.client.Upsert(ctx, authToken, "user_blocks", q, payload, &rows); err != nil {
		return nil, err
	}
	return firstMap(rows, payload), nil
}

func (r *SupabaseRepository) UnblockUser(ctx context.Context, authToken string, relation UserRelation) error {
	q := url.Values{}
	q.Set("blocker_user_id", "eq."+relation.ActorUserID)
	q.Set("blocked_user_id", "eq."+relation.TargetUserID)
	var ignored []map[string]any
	return r.client.Delete(ctx, authToken, "user_blocks", q, &ignored)
}

func (r *SupabaseRepository) MuteUser(ctx context.Context, authToken string, relation UserRelation) (map[string]any, error) {
	payload := map[string]any{"muter_user_id": relation.ActorUserID, "muted_user_id": relation.TargetUserID}
	q := url.Values{}
	q.Set("on_conflict", "muter_user_id,muted_user_id")
	var rows []map[string]any
	if err := r.client.Upsert(ctx, authToken, "user_mutes", q, payload, &rows); err != nil {
		return nil, err
	}
	return firstMap(rows, payload), nil
}

func (r *SupabaseRepository) UnmuteUser(ctx context.Context, authToken string, relation UserRelation) error {
	q := url.Values{}
	q.Set("muter_user_id", "eq."+relation.ActorUserID)
	q.Set("muted_user_id", "eq."+relation.TargetUserID)
	var ignored []map[string]any
	return r.client.Delete(ctx, authToken, "user_mutes", q, &ignored)
}

func (r *SupabaseRepository) HideDrinkLog(ctx context.Context, authToken string, hidden HiddenDrinkLog) (map[string]any, error) {
	payload := map[string]any{"user_id": hidden.UserID, "drink_log_id": hidden.DrinkLogID}
	q := url.Values{}
	q.Set("on_conflict", "user_id,drink_log_id")
	var rows []map[string]any
	if err := r.client.Upsert(ctx, authToken, "feed_hidden_drink_logs", q, payload, &rows); err != nil {
		return nil, err
	}
	return firstMap(rows, payload), nil
}

func (r *SupabaseRepository) UnhideDrinkLog(ctx context.Context, authToken string, hidden HiddenDrinkLog) error {
	q := url.Values{}
	q.Set("user_id", "eq."+hidden.UserID)
	q.Set("drink_log_id", "eq."+hidden.DrinkLogID)
	var ignored []map[string]any
	return r.client.Delete(ctx, authToken, "feed_hidden_drink_logs", q, &ignored)
}

func firstMap(rows []map[string]any, fallback map[string]any) map[string]any {
	if len(rows) == 0 {
		return fallback
	}
	return rows[0]
}
