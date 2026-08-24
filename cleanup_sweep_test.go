package main

import (
	"testing"
	"time"
)

func TestNeedsFollowUpOnlyForDeliveredExpiredMatter(t *testing.T) {
	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	base := matter{ID: "matter-104", SignedAt: now.Add(-48 * time.Hour), DeliveredAt: now.Add(-24 * time.Hour), Deadline: now.Add(-time.Hour)}
	if !needsFollowUp(base, now) {
		t.Fatal("expired delivered matter should be queued")
	}
	base.FollowUpSentAt = now.Add(-time.Minute)
	if needsFollowUp(base, now) {
		t.Fatal("matter with a sent follow-up should not be queued twice")
	}
	base.FollowUpSentAt = time.Time{}
	base.DeliveredAt = time.Time{}
	if needsFollowUp(base, now) {
		t.Fatal("unsigned delivery should not be queued")
	}
}
