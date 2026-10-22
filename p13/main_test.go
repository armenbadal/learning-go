package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchReturnsResponseBeforeDeadline(t *testing.T) {
	const body = "response body"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	got, err := fetch(context.Background(), server.URL, time.Second)
	if err != nil {
		t.Fatalf("fetch() error = %v, want nil", err)
	}
	if got != body {
		t.Fatalf("fetch() = %q, want %q", got, body)
	}
}

func TestFetchReturnsCanceledWhenCallerCancels(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := fetch(ctx, server.URL, time.Second)
		result <- err
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("server did not receive request")
	}
	cancel()

	select {
	case err := <-result:
		if err != context.Canceled {
			t.Fatalf("fetch() error = %v, want %v", err, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Fatal("fetch() did not return after caller canceled")
	}
}

func TestFetchReturnsDeadlineExceededWhenTimeoutExpires(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	_, err := fetch(context.Background(), server.URL, 50*time.Millisecond)
	if err != context.DeadlineExceeded {
		t.Fatalf("fetch() error = %v, want %v", err, context.DeadlineExceeded)
	}
}
