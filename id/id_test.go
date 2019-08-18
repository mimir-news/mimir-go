package id_test

import (
	"testing"

	"github.com/mimir-news/mimir-go/id"
)

func TestNewId(t *testing.T) {
	lastID := ""
	testCases := 100
	for i := 0; i < testCases; i++ {
		newID := id.New()
		if lastID == newID {
			t.Errorf("Should not equal:\nLastId: %s\nNewId: %s", lastID, newID)
		}
		lastID = newID
	}
}
