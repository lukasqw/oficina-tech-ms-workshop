package dto

import (
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
)

func TestToInventoryResponse(t *testing.T) {
	now := time.Now()
	inv, err := inventory.ReconstructInventory(
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		100, 20, 5, now, now, nil,
	)
	if err != nil {
		t.Fatalf("ReconstructInventory: %v", err)
	}

	resp := ToInventoryResponse(inv)
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.ID != "11111111-1111-4111-8111-111111111111" {
		t.Errorf("ID: got %v", resp.ID)
	}
	if resp.ProductID != "22222222-2222-4222-8222-222222222222" {
		t.Errorf("ProductID: got %v", resp.ProductID)
	}
	if resp.AvailableQuantity != 100 {
		t.Errorf("AvailableQuantity: got %d, want 100", resp.AvailableQuantity)
	}
	if resp.ReservedQuantity != 20 {
		t.Errorf("ReservedQuantity: got %d, want 20", resp.ReservedQuantity)
	}
	if resp.PendingQuantity != 5 {
		t.Errorf("PendingQuantity: got %d, want 5", resp.PendingQuantity)
	}
	if resp.CreatedAt == "" {
		t.Error("expected non-empty CreatedAt")
	}
	if resp.UpdatedAt == "" {
		t.Error("expected non-empty UpdatedAt")
	}
}
