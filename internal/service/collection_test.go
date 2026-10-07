package service

import (
	"context"
	"testing"

	"github.com/my-pet-projects/collection/internal/model"
)

func TestGetNextAvailableCollectionSlotWithoutBrewery(t *testing.T) {
	t.Parallel()

	service := CollectionService{}
	slot, err := service.GetNextAvailableCollectionSlot(context.Background(), model.Beer{})
	if err != nil {
		t.Fatalf("GetNextAvailableCollectionSlot() error = %v", err)
	}
	if slot != nil {
		t.Fatalf("slot = %#v, want nil", slot)
	}
}
