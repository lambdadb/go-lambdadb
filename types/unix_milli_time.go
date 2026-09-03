package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// UnixMilliTime is a time.Time that marshals to and from JSON as Unix epoch
// milliseconds.
type UnixMilliTime struct {
	time.Time
}

var (
	_ json.Marshaler   = (*UnixMilliTime)(nil)
	_ json.Unmarshaler = (*UnixMilliTime)(nil)
)

// UnixMilliTimeFrom creates a UnixMilliTime from a time.Time.
func UnixMilliTimeFrom(t time.Time) UnixMilliTime {
	return UnixMilliTime{Time: t}
}

// UnixMilliTimeFromMilliseconds creates a UnixMilliTime from Unix epoch
// milliseconds.
func UnixMilliTimeFromMilliseconds(milliseconds int64) UnixMilliTime {
	return UnixMilliTime{Time: time.UnixMilli(milliseconds)}
}

// MarshalJSON encodes the time as Unix epoch milliseconds.
func (t UnixMilliTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.UnixMilli(), 10)), nil
}

// UnmarshalJSON decodes Unix epoch milliseconds into the time.
func (t *UnixMilliTime) UnmarshalJSON(data []byte) error {
	var milliseconds int64
	if err := json.Unmarshal(data, &milliseconds); err != nil {
		return fmt.Errorf("unix millisecond time: %w", err)
	}
	t.Time = time.UnixMilli(milliseconds)
	return nil
}

// String returns the time in RFC3339Nano format.
func (t UnixMilliTime) String() string {
	return t.Time.Format(time.RFC3339Nano)
}
