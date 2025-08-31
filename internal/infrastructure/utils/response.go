package utils

import (
	"time"
)

func TimePtr(t time.Time) *time.Time {
	return &t
}

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}