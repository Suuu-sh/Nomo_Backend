package drinkinvites

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusRejected InviteStatus = "rejected"
)

type ErrorKind string

const (
	ErrorKindInvalidInput ErrorKind = "invalid_input"
	ErrorKindConflict     ErrorKind = "conflict"
	ErrorKindNotFound     ErrorKind = "not_found"
)

type UserError struct {
	Kind    ErrorKind
	Message string
}

func (e UserError) Error() string {
	return e.Message
}

func ErrorKindOf(err error) (ErrorKind, bool) {
	var userErr UserError
	if errors.As(err, &userErr) {
		return userErr.Kind, true
	}
	return "", false
}

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func CleanUUID(value, field string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "", UserError{Kind: ErrorKindInvalidInput, Message: field + " is required"}
	}
	if !uuidPattern.MatchString(trimmed) {
		return "", UserError{Kind: ErrorKindInvalidInput, Message: field + " must be a valid UUID"}
	}
	return trimmed, nil
}

func CleanDateOnlyOrToday(value, field string, now time.Time) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		if now.IsZero() {
			now = time.Now()
		}
		return now.Format(time.DateOnly), nil
	}
	parsed, err := time.Parse(time.DateOnly, trimmed)
	if err != nil {
		return "", UserError{Kind: ErrorKindInvalidInput, Message: field + " must be YYYY-MM-DD"}
	}
	return parsed.Format(time.DateOnly), nil
}

func ValidateNewInvite(fromUserID, toUserID string) error {
	if fromUserID == toUserID {
		return UserError{Kind: ErrorKindInvalidInput, Message: "cannot invite yourself"}
	}
	return nil
}

func NormalizeResponseStatus(value string) (InviteStatus, error) {
	status := InviteStatus(strings.TrimSpace(value))
	switch status {
	case InviteStatusAccepted, InviteStatusRejected:
		return status, nil
	default:
		return "", UserError{Kind: ErrorKindInvalidInput, Message: "status must be accepted or rejected"}
	}
}

func BlockedDailyStatusMessage(status string) string {
	switch strings.TrimSpace(status) {
	case "liver_rest":
		return "相手が休肝日のため今日は誘えません。"
	case "has_plans":
		return "相手に予定があるため今日は誘えません。"
	default:
		return ""
	}
}

func ExistingInviteConflictMessage(status InviteStatus) string {
	if status == InviteStatusAccepted {
		return "今日はもう予約済みです。"
	}
	return "すでに招待中です。"
}

type NewInvite struct {
	FromUserID string
	ToUserID   string
	InviteDate string
}

type ExistingInvite struct {
	ID     string
	Status InviteStatus
}
