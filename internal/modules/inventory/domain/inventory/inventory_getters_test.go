package inventory

import (
	"testing"
	"time"
)

func TestInventory_TimeGetters(t *testing.T) {
	now := time.Now()
	deleted := now.Add(time.Hour)

	inv, err := ReconstructInventory(
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		10, 2, 1, now, now, &deleted,
	)
	if err != nil {
		t.Fatalf("ReconstructInventory: %v", err)
	}

	if !inv.CreatedAt().Equal(now) {
		t.Errorf("CreatedAt: got %v, want %v", inv.CreatedAt(), now)
	}
	if !inv.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt: got %v, want %v", inv.UpdatedAt(), now)
	}
	if inv.DeletedAt() == nil || !inv.DeletedAt().Equal(deleted) {
		t.Errorf("DeletedAt: got %v, want %v", inv.DeletedAt(), deleted)
	}
}

func TestInventory_DeletedAt_Nil(t *testing.T) {
	now := time.Now()
	inv, err := ReconstructInventory(
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		10, 0, 0, now, now, nil,
	)
	if err != nil {
		t.Fatalf("ReconstructInventory: %v", err)
	}
	if inv.DeletedAt() != nil {
		t.Errorf("expected nil DeletedAt, got %v", inv.DeletedAt())
	}
}
