package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// UnixTime is a time.Time that marshals to/from JSON as seconds since the Unix epoch (integer).
// Use it for API fields that use epoch-second timestamps.
type UnixTime struct {
	time.Time
}

var (
	_ json.Marshaler   = (*UnixTime)(nil)
	_ json.Unmarshaler = (*UnixTime)(nil)
)

// UnixTimeFrom creates a UnixTime from a time.Time.
func UnixTimeFrom(t time.Time) UnixTime {
	return UnixTime{Time: t}
}

// UnixTimeFromSeconds creates a UnixTime from seconds since the Unix epoch.
func UnixTimeFromSeconds(sec int64) UnixTime {
	return UnixTime{Time: time.Unix(sec, 0)}
}

// MarshalJSON encodes the time as seconds since the Unix epoch.
func (t UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.Unix(), 10)), nil
}

// UnmarshalJSON decodes an integer (seconds since the Unix epoch) into the time.
func (t *UnixTime) UnmarshalJSON(data []byte) error {
	var sec int64
	if err := json.Unmarshal(data, &sec); err != nil {
		return fmt.Errorf("unix time: %w", err)
	}
	t.Time = time.Unix(sec, 0)
	return nil
}

// String returns the time in RFC3339 format.
func (t UnixTime) String() string {
	return t.Time.Format(time.RFC3339)
}
