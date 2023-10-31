package reservation

import (
	"time"

	"github.com/migregal/bmstu-iu7-ds-lab2/apiserver/core/ports/library"
	"github.com/migregal/bmstu-iu7-ds-lab2/apiserver/core/ports/rating"
)

type Info struct {
	ID        string
	Username  string
	Status    string
	Start     time.Time
	End       time.Time
	BookID    string
	LibraryID string
}

type FullInfo struct {
	ID           string
	Username     string
	Status       string
	Start        time.Time
	End          time.Time
	ReservedBook library.ReservedBook
	Rating       rating.Rating
}
