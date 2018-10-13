package bstore

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestUnmarshal(t *testing.T) {
	var hs []History

	t1, _ := time.Parse("2006-01-02 15:04:05 MST", "2004-12-12 11:11:12 MSK")
	t2, _ := time.Parse("2006-01-02 MST", "2008-12-11 MSK")
	hs = append(hs, History{Date: t1, Loss: 100})
	hs = append(hs, History{Date: t2, Loss: 10})
	b := Marshal(hs)

	if len(b) == 0 {
		t.Fail()
	}
	if fmt.Sprintf("%x", b) != "20fdbb410000000064d02d4049000000000a" {
		t.Errorf("wrong result have %x, want 20fdbb410000000064d02d4049000000000a", b)
	}
}

func TestMarshal(t *testing.T) {
	b, _ := hex.DecodeString("20fdbb410000000064d02d4049000000000a")
	hs := Unmarshal(b)
	ts := make([]time.Time, 2)
	ts[0], _ = time.Parse("2006-01-02 15:04:05 MST", "2004-12-12 11:11:12 MSK")
	ts[1], _ = time.Parse("2006-01-02 MST", "2008-12-11 MSK")
	for i := range hs {
		if !hs[i].Date.Equal(ts[i]) {
			t.Errorf("date is not equal: have %v,need %v", hs[i].Date, ts[i])
		}
	}
}
