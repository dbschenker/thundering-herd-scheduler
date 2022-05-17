package nodestate

import (
	"testing"
	"time"
)

func TestShouldCountPodsIfExpiryIsNil(t *testing.T) {
	m := []PodNodeStateModel{
		{
			key: "qwe",
		},
		{
			key: "asd",
		},
	}

	nonExpiredCount := countNonExpiredByType(m)
	if nonExpiredCount != 2 {
		t.Errorf("Wrong number of non expired entries calculated, expected 2, but got %d", nonExpiredCount)
	}
}

func TestShouldCountPodsIfExpiryIsSetForSome(t *testing.T) {
	past := time.Now().Add(-2 * time.Minute)
	m := []PodNodeStateModel{
		{
			key: "qwe",
		},
		{
			key:    "asd",
			expiry: &past,
		},
		{
			key: "ert",
		},
	}

	nonExpiredCount := countNonExpiredByType(m)
	if nonExpiredCount != 2 {
		t.Errorf("Wrong number of non expired entries calculated, expected 2, but got %d", nonExpiredCount)
	}
}
