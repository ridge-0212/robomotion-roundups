package v1

import (
	"testing"

	rtesting "github.com/robomotionio/robomotion-go/testing"
)

func TestFetchRoundup(t *testing.T) {
	if !hasRealCredentials() {
		t.Skip("No real API credentials available")
	}

	node := &FetchRoundup{}
	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("roundupId", 1)

	err := q.Run()
	if err != nil {
		t.Fatalf("FetchRoundup failed: %v", err)
	}

	state := q.GetOutput("state")
	if state == nil {
		t.Fatal("expected state output")
	}
}

func TestFetchRoundup_InvalidID(t *testing.T) {
	node := &FetchRoundup{}
	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("roundupId", -1)

	err := q.Run()
	// Local validation should reject negative IDs before hitting API
	if err == nil {
		t.Fatal("expected error for negative roundup ID")
	}
}

func TestFetchRoundup_ZeroID(t *testing.T) {
	node := &FetchRoundup{}
	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("roundupId", 0)

	err := q.Run()
	// Local validation should reject zero IDs before hitting API
	if err == nil {
		t.Fatal("expected error for zero roundup ID")
	}
}
