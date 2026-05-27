package drinkinvites

import (
	"context"
	"reflect"
	"testing"
	"time"
)

const (
	testAuthToken = "access-token"
	testUserID    = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	otherUserID   = "bbbbbbbb-cccc-dddd-eeee-ffffffffffff"
	testInviteID  = "33333333-4444-5555-6666-777777777777"
)

type fakeRepository struct {
	calls       []string
	dailyStatus string
	existing    *ExistingInvite
	created     map[string]any
	updated     map[string]any

	createdInvite NewInvite
	updatedInvite struct {
		inviteID        string
		recipientUserID string
		status          InviteStatus
		respondedAt     time.Time
	}
}

func (f *fakeRepository) ListTodayReservations(context.Context, string, string, string) ([]map[string]any, error) {
	f.calls = append(f.calls, "list_today")
	return nil, nil
}

func (f *fakeRepository) ListIncomingPending(context.Context, string, string, string) ([]map[string]any, error) {
	f.calls = append(f.calls, "list_incoming")
	return nil, nil
}

func (f *fakeRepository) ListOutgoingActive(context.Context, string, string, string) ([]map[string]any, error) {
	f.calls = append(f.calls, "list_outgoing")
	return nil, nil
}

func (f *fakeRepository) DailyStatus(context.Context, string, string, string) (string, error) {
	f.calls = append(f.calls, "daily_status")
	return f.dailyStatus, nil
}

func (f *fakeRepository) FindActiveInviteBetweenUsersForDate(context.Context, string, string, string, string) (*ExistingInvite, error) {
	f.calls = append(f.calls, "find_active")
	return f.existing, nil
}

func (f *fakeRepository) CreateInvite(_ context.Context, _ string, invite NewInvite) (map[string]any, error) {
	f.calls = append(f.calls, "create")
	f.createdInvite = invite
	if f.created != nil {
		return f.created, nil
	}
	return map[string]any{
		"id":           testInviteID,
		"from_user_id": invite.FromUserID,
		"to_user_id":   invite.ToUserID,
		"invite_date":  invite.InviteDate,
		"status":       string(InviteStatusPending),
	}, nil
}

func (f *fakeRepository) UpdatePendingInviteStatus(_ context.Context, _ string, inviteID, recipientUserID string, status InviteStatus, respondedAt time.Time) (map[string]any, error) {
	f.calls = append(f.calls, "update")
	f.updatedInvite.inviteID = inviteID
	f.updatedInvite.recipientUserID = recipientUserID
	f.updatedInvite.status = status
	f.updatedInvite.respondedAt = respondedAt
	return f.updated, nil
}

type fakeNotifier struct {
	received int
	accepted int
}

func (f *fakeNotifier) DrinkInviteReceived(context.Context, string, map[string]any) {
	f.received++
}

func (f *fakeNotifier) DrinkInviteAccepted(context.Context, string, map[string]any) {
	f.accepted++
}

func TestCreateDrinkInviteRejectsSelfInviteBeforeRepositoryAccess(t *testing.T) {
	repo := &fakeRepository{}
	usecase := NewUsecase(Dependencies{Repository: repo})

	_, err := usecase.CreateDrinkInvite(context.Background(), CreateInput{
		AuthToken:  testAuthToken,
		FromUserID: testUserID,
		ToUserID:   testUserID,
		InviteDate: "2026-05-23",
	})

	assertUserError(t, err, ErrorKindInvalidInput, "cannot invite yourself")
	if len(repo.calls) != 0 {
		t.Fatalf("repository calls = %v, want none", repo.calls)
	}
}

func TestCreateDrinkInviteBlocksUnavailableDailyStatus(t *testing.T) {
	repo := &fakeRepository{dailyStatus: "liver_rest"}
	usecase := NewUsecase(Dependencies{Repository: repo})

	_, err := usecase.CreateDrinkInvite(context.Background(), CreateInput{
		AuthToken:  testAuthToken,
		FromUserID: testUserID,
		ToUserID:   otherUserID,
		InviteDate: "2026-05-23",
	})

	assertUserError(t, err, ErrorKindConflict, "相手が休肝日のため今日は誘えません。")
	if want := []string{"daily_status"}; !reflect.DeepEqual(repo.calls, want) {
		t.Fatalf("repository calls = %v, want %v", repo.calls, want)
	}
}

func TestCreateDrinkInviteRejectsExistingAcceptedInvite(t *testing.T) {
	repo := &fakeRepository{existing: &ExistingInvite{ID: testInviteID, Status: InviteStatusAccepted}}
	usecase := NewUsecase(Dependencies{Repository: repo})

	_, err := usecase.CreateDrinkInvite(context.Background(), CreateInput{
		AuthToken:  testAuthToken,
		FromUserID: testUserID,
		ToUserID:   otherUserID,
		InviteDate: "2026-05-23",
	})

	assertUserError(t, err, ErrorKindConflict, "今日はもう予約済みです。")
	if want := []string{"daily_status", "find_active"}; !reflect.DeepEqual(repo.calls, want) {
		t.Fatalf("repository calls = %v, want %v", repo.calls, want)
	}
}

func TestCreateDrinkInviteCreatesPendingInviteAndNotifiesRecipient(t *testing.T) {
	repo := &fakeRepository{}
	notifier := &fakeNotifier{}
	usecase := NewUsecase(Dependencies{Repository: repo, Notifier: notifier})

	row, err := usecase.CreateDrinkInvite(context.Background(), CreateInput{
		AuthToken:  testAuthToken,
		FromUserID: testUserID,
		ToUserID:   otherUserID,
		InviteDate: "2026-05-23",
	})
	if err != nil {
		t.Fatalf("CreateDrinkInvite returned error: %v", err)
	}
	if row["status"] != string(InviteStatusPending) {
		t.Fatalf("created status = %#v", row["status"])
	}
	if repo.createdInvite.InviteDate != "2026-05-23" || repo.createdInvite.FromUserID != testUserID || repo.createdInvite.ToUserID != otherUserID {
		t.Fatalf("created invite = %#v", repo.createdInvite)
	}
	if notifier.received != 1 {
		t.Fatalf("received notifications = %d, want 1", notifier.received)
	}
	if want := []string{"daily_status", "find_active", "create"}; !reflect.DeepEqual(repo.calls, want) {
		t.Fatalf("repository calls = %v, want %v", repo.calls, want)
	}
}

func TestUpdateDrinkInviteAcceptedNotifiesRequester(t *testing.T) {
	respondedAt := time.Date(2026, 5, 23, 12, 34, 56, 0, time.FixedZone("JST", 9*60*60))
	repo := &fakeRepository{updated: map[string]any{"id": testInviteID, "status": string(InviteStatusAccepted)}}
	notifier := &fakeNotifier{}
	usecase := NewUsecase(Dependencies{
		Repository: repo,
		Notifier:   notifier,
		Now:        func() time.Time { return respondedAt },
	})

	row, err := usecase.UpdateDrinkInvite(context.Background(), UpdateInput{
		AuthToken:       testAuthToken,
		InviteID:        testInviteID,
		RecipientUserID: testUserID,
		Status:          "accepted",
	})
	if err != nil {
		t.Fatalf("UpdateDrinkInvite returned error: %v", err)
	}
	if row["id"] != testInviteID {
		t.Fatalf("updated row = %#v", row)
	}
	if repo.updatedInvite.status != InviteStatusAccepted || repo.updatedInvite.inviteID != testInviteID || repo.updatedInvite.recipientUserID != testUserID {
		t.Fatalf("updated invite args = %#v", repo.updatedInvite)
	}
	if got := repo.updatedInvite.respondedAt; !got.Equal(respondedAt.UTC()) {
		t.Fatalf("respondedAt = %s, want %s", got, respondedAt.UTC())
	}
	if notifier.accepted != 1 {
		t.Fatalf("accepted notifications = %d, want 1", notifier.accepted)
	}
}

func TestUpdateDrinkInviteRejectedDoesNotNotify(t *testing.T) {
	repo := &fakeRepository{updated: map[string]any{"id": testInviteID, "status": string(InviteStatusRejected)}}
	notifier := &fakeNotifier{}
	usecase := NewUsecase(Dependencies{Repository: repo, Notifier: notifier})

	_, err := usecase.UpdateDrinkInvite(context.Background(), UpdateInput{
		AuthToken:       testAuthToken,
		InviteID:        testInviteID,
		RecipientUserID: testUserID,
		Status:          "rejected",
	})
	if err != nil {
		t.Fatalf("UpdateDrinkInvite returned error: %v", err)
	}
	if notifier.accepted != 0 {
		t.Fatalf("accepted notifications = %d, want 0", notifier.accepted)
	}
}

func TestUpdateDrinkInviteReturnsNotFoundWhenRepositoryUpdatesNoRows(t *testing.T) {
	repo := &fakeRepository{}
	usecase := NewUsecase(Dependencies{Repository: repo})

	_, err := usecase.UpdateDrinkInvite(context.Background(), UpdateInput{
		AuthToken:       testAuthToken,
		InviteID:        testInviteID,
		RecipientUserID: testUserID,
		Status:          "accepted",
	})

	assertUserError(t, err, ErrorKindNotFound, "drink invite not found")
}

func assertUserError(t *testing.T, err error, wantKind ErrorKind, wantMessage string) {
	t.Helper()
	if err == nil {
		t.Fatal("err = nil")
	}
	kind, ok := ErrorKindOf(err)
	if !ok {
		t.Fatalf("err = %T %v, want UserError", err, err)
	}
	if kind != wantKind || err.Error() != wantMessage {
		t.Fatalf("err = (%s, %q), want (%s, %q)", kind, err.Error(), wantKind, wantMessage)
	}
}
