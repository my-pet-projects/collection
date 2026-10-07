package recognition

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestUserMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		err      error
		expected string
	}{
		"timeout": {
			err:      &ProviderTimeoutError{After: 25 * time.Second, Err: context.DeadlineExceeded},
			expected: "within 25 seconds",
		},
		"quota": {
			err:      &QuotaExceededError{RetryAfter: 2 * time.Minute, Err: errors.New("quota")},
			expected: "Try again in 2m0s",
		},
		"busy": {
			err:      ErrProviderBusy,
			expected: "AI model is busy",
		},
		"unknown": {
			err:      errors.New("provider failure"),
			expected: "temporarily unavailable",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			message := UserMessage(testCase.err)
			if !strings.Contains(message, testCase.expected) {
				t.Fatalf("UserMessage() = %q, want it to contain %q", message, testCase.expected)
			}
		})
	}
}
