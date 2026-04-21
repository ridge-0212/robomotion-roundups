package v1

import (
	"os"
	"testing"

	rtesting "github.com/robomotionio/robomotion-go/testing"
)

var credStore *rtesting.CredentialStore

func TestMain(m *testing.M) {
	credStore = rtesting.NewCredentialStore()
	// Always set a mock credential so validation tests can run without real API keys
	credStore.SetAPIKey("api_key", "test-key-for-validation-tests")
	rtesting.LoadDotEnv(".env")
	credStore.LoadFromEnv("ROUNDUPS", "api_key")
	rtesting.InitCredentials(credStore)

	code := m.Run()

	rtesting.ClearCredentials()
	os.Exit(code)
}

// hasRealCredentials checks if a real API key was loaded from env (not just the mock).
func hasRealCredentials() bool {
	key := os.Getenv("ROUNDUPS_API_KEY")
	return key != ""
}

func TestParseStringArray_StringSlice(t *testing.T) {
	result, err := parseStringArray([]string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Fatalf("expected [a b c], got %v", result)
	}
}

func TestParseStringArray_StringSliceWithEmpty(t *testing.T) {
	result, err := parseStringArray([]string{"a", "", "c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != "a" || result[1] != "c" {
		t.Fatalf("expected [a c], got %v", result)
	}
}

func TestParseStringArray_InterfaceSlice(t *testing.T) {
	result, err := parseStringArray([]interface{}{"a", "b", "c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
}

func TestParseStringArray_InterfaceSliceWithNonString(t *testing.T) {
	_, err := parseStringArray([]interface{}{"a", 42, "c"})
	if err == nil {
		t.Fatal("expected error for non-string element")
	}
}

func TestParseStringArray_SingleString(t *testing.T) {
	result, err := parseStringArray("single")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "single" {
		t.Fatalf("expected [single], got %v", result)
	}
}

func TestParseStringArray_EmptyString(t *testing.T) {
	result, err := parseStringArray("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil for empty string, got %v", result)
	}
}

func TestParseStringArray_UnsupportedType(t *testing.T) {
	_, err := parseStringArray(123)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
