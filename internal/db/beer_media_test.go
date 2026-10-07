package db

import (
	"context"
	"testing"

	"github.com/my-pet-projects/collection/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeleteBeerMediaKeepsSharedMediaItem(t *testing.T) {
	store, database := newBeerMediaTestStore(t)
	createBeerMediaTestRecords(t, database)

	mediaItemDeleted, err := store.DeleteBeerMedia(context.Background(), model.BeerMedia{ID: 1, MediaID: 10})
	if err != nil {
		t.Fatalf("DeleteBeerMedia() error = %v", err)
	}
	if mediaItemDeleted {
		t.Fatal("DeleteBeerMedia() deleted a referenced media item")
	}

	assertBeerMediaTestCount(t, database, &model.BeerMedia{}, 1)
	assertBeerMediaTestCount(t, database, &model.MediaItem{}, 1)
}

func TestDeleteBeerMediaDeletesUnreferencedMediaItem(t *testing.T) {
	store, database := newBeerMediaTestStore(t)
	createBeerMediaTestRecords(t, database)

	mediaItemDeleted, err := store.DeleteBeerMedia(context.Background(), model.BeerMedia{ID: 1, MediaID: 10})
	if err != nil {
		t.Fatalf("DeleteBeerMedia() first error = %v", err)
	}
	if mediaItemDeleted {
		t.Fatal("DeleteBeerMedia() first call deleted a referenced media item")
	}

	mediaItemDeleted, err = store.DeleteBeerMedia(context.Background(), model.BeerMedia{ID: 2, MediaID: 10})
	if err != nil {
		t.Fatalf("DeleteBeerMedia() second error = %v", err)
	}
	if !mediaItemDeleted {
		t.Fatal("DeleteBeerMedia() did not delete an unreferenced media item")
	}

	assertBeerMediaTestCount(t, database, &model.BeerMedia{}, 0)
	assertBeerMediaTestCount(t, database, &model.MediaItem{}, 0)
}

func newBeerMediaTestStore(t *testing.T) (BeerMediaStore, *gorm.DB) {
	t.Helper()

	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err = database.AutoMigrate(&model.MediaItem{}, &model.BeerMedia{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	return BeerMediaStore{db: &DbClient{gorm: database}}, database
}

func createBeerMediaTestRecords(t *testing.T, database *gorm.DB) {
	t.Helper()

	media := model.MediaItem{ID: 10, Hash: "shared-media"}
	if err := database.Create(&media).Error; err != nil {
		t.Fatalf("create media item: %v", err)
	}
	items := []model.BeerMedia{
		{ID: 1, MediaID: media.ID, Type: model.BeerMediaBottle},
		{ID: 2, MediaID: media.ID, Type: model.BeerMediaBottle},
	}
	if err := database.Create(&items).Error; err != nil {
		t.Fatalf("create beer media items: %v", err)
	}
}

func assertBeerMediaTestCount(t *testing.T, database *gorm.DB, value any, want int64) {
	t.Helper()

	var got int64
	if err := database.Model(value).Count(&got).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	if got != want {
		t.Fatalf("record count = %d, want %d", got, want)
	}
}
