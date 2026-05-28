package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/yota/nomo/backend/internal/features/notifications"
)

func (r *router) notificationUsecase(_ *http.Request) *notifications.Usecase {
	return notifications.NewUsecase(notifications.Dependencies{
		Repository: notifications.NewSupabaseRepository(r.deps.Supabase, r.deps.AdminSupabase, r.deps.Config.SupabaseServiceRoleKey),
		PushSender: r.deps.FCM,
		Logger:     r.deps.Logger,
	})
}

func (r *router) createFriendRequestReceivedNotification(req *http.Request, authToken string, requestRow map[string]any) {
	r.notificationUsecase(req).NotifyFriendRequestReceived(req.Context(), authToken, requestRow)
}

func (r *router) createFriendRequestAcceptedNotification(req *http.Request, authToken string, requestRow map[string]any) {
	r.notificationUsecase(req).NotifyFriendRequestAccepted(req.Context(), authToken, requestRow)
}

func (r *router) createDrinkInviteReceivedNotification(req *http.Request, authToken string, inviteRow map[string]any) {
	r.notificationUsecase(req).NotifyDrinkInviteReceived(req.Context(), authToken, inviteRow)
}

func (r *router) createDrinkInviteAcceptedNotification(req *http.Request, authToken string, inviteRow map[string]any) {
	r.notificationUsecase(req).NotifyDrinkInviteAccepted(req.Context(), authToken, inviteRow)
}

func (r *router) createDrinkLogTaggedNotifications(req *http.Request, authToken, logID, ownerUserID string, friendIDs []string) {
	r.notificationUsecase(req).NotifyDrinkLogTagged(req.Context(), authToken, logID, ownerUserID, friendIDs)
}

func (r *router) createDrinkLogLikeNotification(req *http.Request, authToken, logID, actorUserID string) {
	r.notificationUsecase(req).NotifyDrinkLogLiked(req.Context(), authToken, logID, actorUserID)
}

type notificationOutboxEvent struct {
	EventKind       string
	AggregateType   string
	AggregateID     string
	ActorUserID     string
	RecipientUserID string
	Payload         map[string]any
}

func (r *router) recordNotificationOutboxEvent(ctx context.Context, event notificationOutboxEvent) {
	if r.deps.AdminSupabase == nil || r.deps.Config.SupabaseServiceRoleKey == "" || event.EventKind == "" {
		return
	}
	payload := event.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	body := map[string]any{
		"event_kind":     event.EventKind,
		"aggregate_type": event.AggregateType,
		"payload":        payload,
		"status":         "processed",
		"attempts":       1,
		"processed_at":   time.Now().UTC().Format(time.RFC3339),
	}
	if event.AggregateID != "" {
		body["aggregate_id"] = event.AggregateID
	}
	if event.ActorUserID != "" {
		body["actor_user_id"] = event.ActorUserID
	}
	if event.RecipientUserID != "" {
		body["recipient_user_id"] = event.RecipientUserID
	}
	var ignored []map[string]any
	if err := r.deps.AdminSupabase.Post(ctx, r.deps.Config.SupabaseServiceRoleKey, "notification_outbox", nil, body, &ignored); err != nil && r.deps.Logger != nil {
		r.deps.Logger.Warn("failed to record notification outbox event", "event", event.EventKind, "error", err)
	}
}

func (r *router) adminCreateNotification(w http.ResponseWriter, req *http.Request, _ AuthUser) {
	var input AdminCreateSystemNotificationRequest
	if !decodeJSONBody(w, req, &input) {
		return
	}
	result, err := r.notificationUsecase(req).CreateSystemNotifications(req.Context(), notifications.CreateSystemInput{
		Title:            input.Title,
		Message:          input.Message,
		RecipientUserIDs: input.RecipientUserIDs,
		SendToAll:        input.SendToAll,
		SystemKey:        input.SystemKey,
	})
	if err != nil {
		writeNotificationError(w, err)
		return
	}
	r.recordNotificationOutboxEvent(req.Context(), notificationOutboxEvent{
		EventKind:     "system_notification.created",
		AggregateType: "system_notification",
		Payload: map[string]any{
			"title":              input.Title,
			"message":            input.Message,
			"recipient_user_ids": input.RecipientUserIDs,
			"send_to_all":        input.SendToAll,
			"system_key":         input.SystemKey,
			"recipient_count":    result.RecipientCount,
			"created_count":      result.CreatedCount,
		},
	})
	writeJSON(w, http.StatusCreated, result)
}

func writeNotificationError(w http.ResponseWriter, err error) {
	if kind, ok := notifications.ErrorKindOf(err); ok {
		switch kind {
		case notifications.ErrorKindInvalidInput:
			writeError(w, http.StatusBadRequest, err.Error())
		case notifications.ErrorKindUpstream:
			writeError(w, http.StatusBadGateway, "upstream service error")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeSupabaseError(w, err)
}
