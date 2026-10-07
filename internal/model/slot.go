package model

import (
	"errors"
	"fmt"
)

const (
	maxCollectionSheets = 100
	columnsPerSheet     = 7
)

// FindFirstAvailableSlot returns the first unoccupied collection slot.
func FindFirstAvailableSlot(occupiedSlotIDs []string, prefix string, rowsPerSheet int) (Slot, error) {
	if rowsPerSheet != SmallSheetRows && rowsPerSheet != LargeSheetRows {
		return Slot{}, fmt.Errorf("invalid rows per sheet: %d", rowsPerSheet)
	}

	occupied := make(map[string]struct{}, len(occupiedSlotIDs))
	for _, slotID := range occupiedSlotIDs {
		occupied[slotID] = struct{}{}
	}

	current := NewFirstSlot(prefix)
	maxSlots := maxCollectionSheets * columnsPerSheet * rowsPerSheet
	for range maxSlots {
		if _, exists := occupied[current.String()]; !exists {
			return current, nil
		}
		current = current.NextSlot(rowsPerSheet)
	}

	return Slot{}, errors.New("no collection slots available")
}
