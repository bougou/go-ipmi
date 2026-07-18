package client

import "testing"

func TestV15SeedInSeq(t *testing.T) {
	tests := []struct {
		starting uint32
		want     uint32
	}{
		{0, 0}, // reserved 0 → treat as 1 → seed 0 (first packet 1)
		{1, 0}, // start 1 → seed 0
		{0xf5f6fd46, 0xf5f6fd45},
		{0xffffffff, 0xfffffffe},
	}
	for _, tc := range tests {
		if got := v15SeedInSeq(tc.starting); got != tc.want {
			t.Errorf("v15SeedInSeq(0x%08x)=0x%08x, want 0x%08x", tc.starting, got, tc.want)
		}
	}
}

func TestV15InSeqIncrementSkipsZero(t *testing.T) {
	c := &Client{session: &session{}}
	c.session.v15.active = true
	c.session.v15.inSeq = 0xffffffff
	sess, err := c.genSession15([]byte{0x20, 0x18, 0xc8, 0x81, 0x00, 0x01, 0x00})
	if err != nil {
		t.Fatalf("genSession15: %v", err)
	}
	if sess.SessionHeader15.Sequence != 1 {
		t.Fatalf("wrap sequence: want 1, got 0x%08x", sess.SessionHeader15.Sequence)
	}
	if c.session.v15.inSeq != 1 {
		t.Fatalf("inSeq after wrap: want 1, got 0x%08x", c.session.v15.inSeq)
	}
}
