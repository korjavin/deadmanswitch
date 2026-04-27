package templates

import (
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

func TestHeartbeatStatusOk(t *testing.T) {
	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	user := &models.User{
		LastActivity:  now.AddDate(0, 0, -1),
		PingFrequency: 7,
		PingDeadline:  14,
	}
	variant, label := HeartbeatStatus(user, now)
	if variant != "ok" || label != "OK" {
		t.Errorf("expected ok/OK, got %s/%s", variant, label)
	}
}

func TestHeartbeatStatusWarnPastNudge(t *testing.T) {
	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	user := &models.User{
		LastActivity:  now.AddDate(0, 0, -8), // past 7d nudge, before 14d deadline
		PingFrequency: 7,
		PingDeadline:  14,
	}
	variant, label := HeartbeatStatus(user, now)
	if variant != "warn" || label != "OVERDUE" {
		t.Errorf("expected warn/OVERDUE, got %s/%s", variant, label)
	}
}

func TestHeartbeatStatusCritPastDeadline(t *testing.T) {
	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	user := &models.User{
		LastActivity:  now.AddDate(0, 0, -15), // past 14d deadline
		PingFrequency: 7,
		PingDeadline:  14,
	}
	variant, label := HeartbeatStatus(user, now)
	if variant != "crit" || label != "URGENT" {
		t.Errorf("expected crit/URGENT, got %s/%s", variant, label)
	}
}

func TestSetHeartbeatNilUserNoOp(t *testing.T) {
	d := TemplateData{HeartbeatVariant: "warn", HeartbeatLabel: "OVERDUE"}
	d.SetHeartbeat(nil)
	if d.HeartbeatVariant != "warn" || d.HeartbeatLabel != "OVERDUE" {
		t.Errorf("nil user must not overwrite existing values; got %s/%s", d.HeartbeatVariant, d.HeartbeatLabel)
	}
}
