package model

import "testing"

func TestCountrySlotConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  CountrySlotConfig
		wantErr bool
	}{
		{
			name:    "valid shared prefix",
			config:  CountrySlotConfig{Prefix: "GBR/IRL", RowsPerSheet: LargeSheetRows},
			wantErr: false,
		},
		{
			name:    "valid short sheet",
			config:  CountrySlotConfig{Prefix: "MIDE", RowsPerSheet: SmallSheetRows},
			wantErr: false,
		},
		{
			name:    "invalid separator",
			config:  CountrySlotConfig{Prefix: "GBR-IRL", RowsPerSheet: LargeSheetRows},
			wantErr: true,
		},
		{
			name:    "invalid row count",
			config:  CountrySlotConfig{Prefix: "GBR", RowsPerSheet: 7},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, hasErrs := test.config.Validate()
			if hasErrs != test.wantErr {
				t.Fatalf("Validate() hasErrs = %v, want %v", hasErrs, test.wantErr)
			}
		})
	}
}

func TestSlotNextSlotUsesConfiguredRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		slot         Slot
		rowsPerSheet int
		want         Slot
	}{
		{
			name:         "five row sheet advances column",
			slot:         Slot{GeoPrefix: "MIDE", SheetID: "C1", SheetSlot: "A5"},
			rowsPerSheet: SmallSheetRows,
			want:         Slot{GeoPrefix: "MIDE", SheetID: "C1", SheetSlot: "B1"},
		},
		{
			name:         "six row sheet stays in column",
			slot:         Slot{GeoPrefix: "DEU", SheetID: "C1", SheetSlot: "A5"},
			rowsPerSheet: LargeSheetRows,
			want:         Slot{GeoPrefix: "DEU", SheetID: "C1", SheetSlot: "A6"},
		},
		{
			name:         "last position advances sheet",
			slot:         Slot{GeoPrefix: "DEU", SheetID: "C3", SheetSlot: "G6"},
			rowsPerSheet: LargeSheetRows,
			want:         Slot{GeoPrefix: "DEU", SheetID: "C4", SheetSlot: "A1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.slot.NextSlot(test.rowsPerSheet); got != test.want {
				t.Fatalf("NextSlot(%d) = %#v, want %#v", test.rowsPerSheet, got, test.want)
			}
		})
	}
}

func TestFindFirstAvailableSlot(t *testing.T) {
	t.Parallel()

	t.Run("returns first slot when collection is empty", func(t *testing.T) {
		t.Parallel()

		got, err := FindFirstAvailableSlot(nil, "DEU", LargeSheetRows)
		if err != nil {
			t.Fatalf("FindFirstAvailableSlot() error = %v", err)
		}
		want := Slot{GeoPrefix: "DEU", SheetID: "C1", SheetSlot: "A1"}
		if got != want {
			t.Fatalf("FindFirstAvailableSlot() = %#v, want %#v", got, want)
		}
	})

	t.Run("returns first gap", func(t *testing.T) {
		t.Parallel()

		occupied := []string{"DEU-C1-A1", "DEU-C1-A2", "DEU-C1-A4"}
		got, err := FindFirstAvailableSlot(occupied, "DEU", LargeSheetRows)
		if err != nil {
			t.Fatalf("FindFirstAvailableSlot() error = %v", err)
		}
		want := Slot{GeoPrefix: "DEU", SheetID: "C1", SheetSlot: "A3"}
		if got != want {
			t.Fatalf("FindFirstAvailableSlot() = %#v, want %#v", got, want)
		}
	})

	t.Run("rejects invalid row count", func(t *testing.T) {
		t.Parallel()

		_, err := FindFirstAvailableSlot(nil, "DEU", 7)
		if err == nil {
			t.Fatal("FindFirstAvailableSlot() error = nil, want error")
		}
	})
}
