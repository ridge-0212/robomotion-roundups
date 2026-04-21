package v1

import (
	"testing"

	rtesting "github.com/robomotionio/robomotion-go/testing"
)

func TestCreateRoundup_Minimal(t *testing.T) {
	if !hasRealCredentials() {
		t.Skip("No real API credentials available")
	}

	node := &CreateRoundup{}
	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("headline", "Best office chairs for 2026")

	err := q.Run()
	if err != nil {
		t.Fatalf("CreateRoundup failed: %v", err)
	}

	roundupID := q.GetOutput("roundupId")
	if roundupID == nil {
		t.Fatal("expected roundupId output")
	}

	state := q.GetOutput("state")
	if state == nil {
		t.Fatal("expected state output")
	}
}

func TestCreateRoundup_WithOptions(t *testing.T) {
	if !hasRealCredentials() {
		t.Skip("No real API credentials available")
	}

	node := &CreateRoundup{}
	// Set raw option fields directly since they are not variable wrappers
	node.OptToneOfVoice = "Informative"
	node.OptLanguage = "English"

	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("headline", "Top 5 standing desks").
		SetInput("keywords", "standing desk, ergonomic")

	err := q.Run()
	if err != nil {
		t.Fatalf("CreateRoundup with options failed: %v", err)
	}

	roundupID := q.GetOutput("roundupId")
	if roundupID == nil {
		t.Fatal("expected roundupId output")
	}
}

func TestCreateRoundup_MissingRequired(t *testing.T) {
	node := &CreateRoundup{}
	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key")
	// No headline, targetAudience, or keywords provided

	err := q.Run()
	if err == nil {
		t.Fatal("expected error when no required parameters provided")
	}
}

func TestCreateRoundup_NegativeProductCoverCount(t *testing.T) {
	node := &CreateRoundup{}
	node.OptCoverImageStyle = "product"
	node.OptProductCoverCount = -1

	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("headline", "Test")

	err := q.Run()
	if err == nil {
		t.Fatal("expected error for negative product cover count")
	}
}

func TestCreateRoundup_CoverCountWithoutProductStyle(t *testing.T) {
	node := &CreateRoundup{}
	node.OptProductCoverCount = 3
	// OptCoverImageStyle not set to "product"

	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("headline", "Test")

	err := q.Run()
	if err == nil {
		t.Fatal("expected error when product cover count is set without product cover image style")
	}
}

func TestCreateRoundup_UnifiedWithoutURLs(t *testing.T) {
	node := &CreateRoundup{}
	node.OptProductType = "unified"

	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("headline", "Test")

	err := q.Run()
	if err == nil {
		t.Fatal("expected error when product type is unified but no URLs provided")
	}
}

func TestCreateRoundup_ExceedsProductsCount(t *testing.T) {
	node := &CreateRoundup{}
	node.OptProductsCount = 51

	q := rtesting.NewQuick(node).
		SetCredential("OptAPIKey", "api_key", "api_key").
		SetInput("headline", "Test")

	err := q.Run()
	if err == nil {
		t.Fatal("expected error when products count exceeds 50")
	}
}
