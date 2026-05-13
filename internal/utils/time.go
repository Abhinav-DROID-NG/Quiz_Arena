package utils

import "time"

// NowUTC returns the current UTC time.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// ElapsedSeconds returns the number of whole seconds elapsed since t.
func ElapsedSeconds(t time.Time) int {
	return int(time.Since(t).Seconds())
}

// ElapsedMillis returns the number of milliseconds elapsed since t.
func ElapsedMillis(t time.Time) int64 {
	return time.Since(t).Milliseconds()
}

// DeadlineExceeded returns true if the given start time plus duration has passed.
func DeadlineExceeded(start time.Time, limitSeconds int) bool {
	return time.Since(start) > time.Duration(limitSeconds)*time.Second
}
