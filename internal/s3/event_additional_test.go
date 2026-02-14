package s3

import (
	"errors"
	"testing"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"github.com/minio/minio-go/v7/pkg/notification"
)

func TestParseEventType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected types.EventType
	}{
		{name: "created", input: "s3:ObjectCreated:Put", expected: types.EventCreated},
		{name: "removed", input: "s3:ObjectRemoved:Delete", expected: types.EventRemoved},
		{name: "unknown", input: "s3:ObjectAccessed:Get", expected: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := parseEventType(tc.input)
			if got != tc.expected {
				t.Fatalf("unexpected event type: got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestEnsureNoError(t *testing.T) {
	t.Parallel()

	t.Run("all channels empty", func(t *testing.T) {
		t.Parallel()

		ch := make(chan notification.Info)
		if err := ensureNoError(ch); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("channel closed", func(t *testing.T) {
		t.Parallel()

		ch := make(chan notification.Info)
		close(ch)

		err := ensureNoError(ch)
		if !errors.Is(err, errChanClosed) {
			t.Fatalf("expected errChanClosed, got %v", err)
		}
	})

	t.Run("channel with notification error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("boom")

		ch := make(chan notification.Info, 1)
		ch <- notification.Info{Err: wantErr}

		err := ensureNoError(ch)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})
}

func TestParseEventTime(t *testing.T) {
	t.Parallel()

	const raw = "2024-02-03T04:05:06.000Z"

	tm := parseEventTime(raw)
	if tm.UTC().Format("2006-01-02T15:04:05.000Z") != raw {
		t.Fatalf("unexpected parsed time: %s", tm.UTC().Format("2006-01-02T15:04:05.000Z"))
	}

	before := time.Now()
	got := parseEventTime("not-a-date")
	after := time.Now()

	if got.Before(before.Add(-time.Second)) || got.After(after.Add(time.Second)) {
		t.Fatalf("expected fallback time near now, got %v", got)
	}
}
