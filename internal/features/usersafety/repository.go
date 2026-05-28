package usersafety

import "context"

type Repository interface {
	BlockUser(ctx context.Context, authToken string, relation UserRelation) (map[string]any, error)
	UnblockUser(ctx context.Context, authToken string, relation UserRelation) error
	MuteUser(ctx context.Context, authToken string, relation UserRelation) (map[string]any, error)
	UnmuteUser(ctx context.Context, authToken string, relation UserRelation) error
	HideDrinkLog(ctx context.Context, authToken string, hidden HiddenDrinkLog) (map[string]any, error)
	UnhideDrinkLog(ctx context.Context, authToken string, hidden HiddenDrinkLog) error
	CleanupBlockedRelations(ctx context.Context, relation UserRelation) error
}
