package mockstore

import (
	"context"
	"testing"
)

func TestMockStoreRoundTrip(t *testing.T) {
	ms := New()
	ctx := context.Background()

	if err := ms.Put(ctx, "b", "k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	val, err := ms.Get(ctx, "b", "k")
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "v" {
		t.Errorf("expected v, got %s", string(val))
	}

	calls := ms.Calls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].Method != "Put" {
		t.Errorf("expected Put, got %s", calls[0].Method)
	}
}

func TestMockStoreErrorInjection(t *testing.T) {
	ms := New()
	ms.SetError("Get", context.DeadlineExceeded)

	_, err := ms.Get(context.Background(), "b", "k")
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestAssertPutCalled(t *testing.T) {
	ms := New()
	ms.Put(context.Background(), "bucket", "key", []byte("val"))
	AssertPutCalled(t, ms, "bucket", "key")
}
