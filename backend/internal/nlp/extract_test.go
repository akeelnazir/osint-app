package nlp

import (
	"testing"
)

func TestExtractEmails(t *testing.T) {
	entities := Extract("Contact me at john@example.com or admin@bellingcat.com")
	emailCount := 0
	for _, e := range entities {
		if e.Type == "other" && contains(e.Value, "@") {
			emailCount++
		}
	}
	if emailCount != 2 {
		t.Errorf("expected 2 email entities, got %d", emailCount)
	}
}

func TestExtractURLs(t *testing.T) {
	entities := Extract("Visit https://example.com and http://bellingcat.com")
	urlCount := 0
	for _, e := range entities {
		if e.Type == "other" && (startsWith(e.Value, "http://") || startsWith(e.Value, "https://")) {
			urlCount++
		}
	}
	if urlCount != 2 {
		t.Errorf("expected 2 URL entities, got %d", urlCount)
	}
}

func TestExtractOrganizations(t *testing.T) {
	entities := Extract("The UN and NATO issued a joint statement with the EU.")
	orgCount := 0
	for _, e := range entities {
		if e.Type == "organization" {
			orgCount++
		}
	}
	if orgCount < 3 {
		t.Errorf("expected at least 3 organization entities (UN, NATO, EU), got %d", orgCount)
	}
}

func TestExtractCountries(t *testing.T) {
	entities := Extract("The conflict in Syria and Ukraine has affected Russia and Ukraine.")
	locCount := 0
	for _, e := range entities {
		if e.Type == "location" {
			locCount++
		}
	}
	if locCount < 3 {
		t.Errorf("expected at least 3 location entities, got %d", locCount)
	}
}

func TestExtractCoordinates(t *testing.T) {
	entities := Extract("The incident occurred at 35.6762, 139.6503 near Tokyo.")
	locCount := 0
	for _, e := range entities {
		if e.Type == "location" && contains(e.Value, ",") {
			locCount++
		}
	}
	if locCount != 1 {
		t.Errorf("expected 1 coordinate entity, got %d", locCount)
	}
}

func TestExtractEmpty(t *testing.T) {
	entities := Extract("")
	if len(entities) != 0 {
		t.Errorf("expected 0 entities for empty input, got %d", len(entities))
	}
}

func TestExtractDeduplication(t *testing.T) {
	entities := Extract("UN UN UN NATO NATO")
	seen := map[string]bool{}
	for _, e := range entities {
		key := e.Type + "|" + e.Value
		if seen[key] {
			t.Errorf("duplicate entity found: %s", key)
		}
		seen[key] = true
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
