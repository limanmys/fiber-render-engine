package constants

import (
	"testing"
	"time"
)

func TestIstanbulLocationUsesLocalOffset(t *testing.T) {
	location := loadIstanbulLocation()
	_, offset := time.Date(2026, time.January, 1, 12, 0, 0, 0, location).Zone()
	if offset != 3*60*60 {
		t.Fatalf("Istanbul offset = %d, want %d", offset, 3*60*60)
	}
}
