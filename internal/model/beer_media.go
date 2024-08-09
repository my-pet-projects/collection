package model

type BeerMedia struct {
	ID      int
	BeerID  *int
	MediaID int
	Type    BeerMediaType
}

type BeerMediaType int

const (
	BeerMediaBottle      BeerMediaType = 1
	BeerMediaLabel       BeerMediaType = 2
	BeerMediaCrownCap    BeerMediaType = 3
	BeerMediaTwistOffCap BeerMediaType = 4
	BeerMediaPullOffCap  BeerMediaType = 5
	BeerMediaCeramicCap  BeerMediaType = 6
)

func NewBeerMedia(mediaItem *MediaItem, content *MediaItemContent) *BeerMedia {
	return &BeerMedia{
		MediaID: mediaItem.ID,
		Type:    1,
	}
}
