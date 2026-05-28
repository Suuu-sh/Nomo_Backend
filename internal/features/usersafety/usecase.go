package usersafety

import "context"

type Dependencies struct {
	Repository Repository
}

type Usecase struct {
	repository Repository
}

func NewUsecase(deps Dependencies) *Usecase {
	return &Usecase{repository: deps.Repository}
}

type UserTargetInput struct {
	AuthToken    string
	ActorUserID  string
	TargetUserID string
}

type DrinkLogInput struct {
	AuthToken  string
	UserID     string
	DrinkLogID string
}

func (u *Usecase) BlockUser(ctx context.Context, input UserTargetInput) (map[string]any, error) {
	relation, err := cleanUserRelation(input)
	if err != nil {
		return nil, err
	}
	return u.repository.BlockUser(ctx, input.AuthToken, relation)
}

func (u *Usecase) UnblockUser(ctx context.Context, input UserTargetInput) error {
	relation, err := cleanUserRelation(input)
	if err != nil {
		return err
	}
	return u.repository.UnblockUser(ctx, input.AuthToken, relation)
}

func (u *Usecase) MuteUser(ctx context.Context, input UserTargetInput) (map[string]any, error) {
	relation, err := cleanUserRelation(input)
	if err != nil {
		return nil, err
	}
	return u.repository.MuteUser(ctx, input.AuthToken, relation)
}

func (u *Usecase) UnmuteUser(ctx context.Context, input UserTargetInput) error {
	relation, err := cleanUserRelation(input)
	if err != nil {
		return err
	}
	return u.repository.UnmuteUser(ctx, input.AuthToken, relation)
}

func (u *Usecase) HideDrinkLog(ctx context.Context, input DrinkLogInput) (map[string]any, error) {
	hidden, err := cleanHiddenDrinkLog(input)
	if err != nil {
		return nil, err
	}
	return u.repository.HideDrinkLog(ctx, input.AuthToken, hidden)
}

func (u *Usecase) UnhideDrinkLog(ctx context.Context, input DrinkLogInput) error {
	hidden, err := cleanHiddenDrinkLog(input)
	if err != nil {
		return err
	}
	return u.repository.UnhideDrinkLog(ctx, input.AuthToken, hidden)
}

func cleanUserRelation(input UserTargetInput) (UserRelation, error) {
	actorUserID, err := CleanUUID(input.ActorUserID, "user id")
	if err != nil {
		return UserRelation{}, err
	}
	targetUserID, err := CleanUUID(input.TargetUserID, "target user id")
	if err != nil {
		return UserRelation{}, err
	}
	if err := ValidateDifferentUsers(actorUserID, targetUserID); err != nil {
		return UserRelation{}, err
	}
	return UserRelation{ActorUserID: actorUserID, TargetUserID: targetUserID}, nil
}

func cleanHiddenDrinkLog(input DrinkLogInput) (HiddenDrinkLog, error) {
	userID, err := CleanUUID(input.UserID, "user id")
	if err != nil {
		return HiddenDrinkLog{}, err
	}
	drinkLogID, err := CleanUUID(input.DrinkLogID, "drink log id")
	if err != nil {
		return HiddenDrinkLog{}, err
	}
	return HiddenDrinkLog{UserID: userID, DrinkLogID: drinkLogID}, nil
}
