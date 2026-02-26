package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUnixTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
		want int64
	}{
		{"seconds", `1700000000`, 1700000000},
		{"zero", `0`, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u UnixTime
			if err := json.Unmarshal([]byte(tt.json), &u); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}
			if got := u.Unix(); got != tt.want {
				t.Errorf("Unix() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestUnixTime_MarshalJSON(t *testing.T) {
	u := UnixTimeFromSeconds(1700000000)
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(data) != "1700000000" {
		t.Errorf("MarshalJSON = %s, want 1700000000", data)
	}
}

func TestUnixTime_RoundTrip(t *testing.T) {
	orig := UnixTimeFrom(time.Date(2023, 11, 15, 12, 0, 0, 0, time.UTC))
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded UnixTime
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Unix() != orig.Unix() {
		t.Errorf("round trip: got unix %d, want %d", decoded.Unix(), orig.Unix())
	}
}

func TestUnixTime_String(t *testing.T) {
	u := UnixTimeFromSeconds(1700000000)
	s := u.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
	_, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Errorf("String() should be RFC3339: %v", err)
	}
}
