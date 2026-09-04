// Package nlp provides basic regex + gazetteer entity extraction for OSINT text.
package nlp

import (
	"regexp"
	"strings"
)

// Entity is an extracted mention.
type Entity struct {
	Type       string  // person | organization | location | other
	Value      string
	Confidence float64
}

// Gazetteers (small, built-in). Extend as needed.
var orgGazetteer = map[string]bool{
	"UN": true, "United Nations": true, "NATO": true, "EU": true, "European Union": true,
	"CIA": true, "FBI": true, "NSA": true, "Interpol": true, "WHO": true,
	"Amnesty International": true, "Red Cross": true, "ICRC": true,
	"Google": true, "Microsoft": true, "Apple": true, "Meta": true, "Facebook": true,
	"Twitter": true, "Bellingcat": true,
}

var countryGazetteer = []string{
	"Afghanistan", "Albania", "Algeria", "Argentina", "Australia", "Austria",
	"Belgium", "Brazil", "Bulgaria", "Canada", "China", "Colombia", "Cuba",
	"Cyprus", "Czech Republic", "Denmark", "Egypt", "Estonia", "Finland",
	"France", "Germany", "Greece", "Hungary", "Iceland", "India", "Iran",
	"Iraq", "Ireland", "Israel", "Italy", "Japan", "Jordan", "Kenya",
	"Korea", "Latvia", "Lebanon", "Libya", "Lithuania", "Luxembourg",
	"Malaysia", "Mexico", "Morocco", "Netherlands", "Nigeria", "Norway",
	"Pakistan", "Poland", "Portugal", "Qatar", "Romania", "Russia",
	"Saudi Arabia", "Serbia", "Singapore", "Slovakia", "Slovenia",
	"South Africa", "Spain", "Sudan", "Sweden", "Switzerland", "Syria",
	"Taiwan", "Thailand", "Tunisia", "Turkey", "Ukraine", "United Kingdom",
	"UK", "USA", "United States", "Venezuela", "Vietnam", "Yemen",
}

var (
	// Email regex.
	emailRe = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	// URL regex.
	urlRe = regexp.MustCompile(`https?://[^\s<>"']+`)
	// IPv4 regex.
	ipv4Re = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	// Coordinates: lat,lng or lat lng pairs.
	coordRe = regexp.MustCompile(`-?\d{1,3}\.\d+,\s*-?\d{1,3}\.\d+`)
	// Hashtag.
	hashtagRe = regexp.MustCompile(`#\w+`)
	// @-mention (treated as "person" candidate).
	mentionRe = regexp.MustCompile(`@[A-Za-z0-9_]+`)
	// All-caps acronyms 2-6 chars (org candidates).
	acronymRe = regexp.MustCompile(`\b[A-Z]{2,6}\b`)
	// Capitalized phrases (person/location candidates): 1-3 Capitalized words.
	capPhraseRe = regexp.MustCompile(`\b(?:[A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\b`)
)

// Extract runs all extractors over the input text and returns deduplicated entities.
func Extract(text string) []Entity {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	seen := map[string]bool{}
	var out []Entity
	add := func(typ, val string, conf float64) {
		val = strings.TrimSpace(val)
		key := typ + "|" + strings.ToLower(val)
		if val == "" || seen[key] {
			return
		}
		seen[key] = true
		out = append(out, Entity{Type: typ, Value: val, Confidence: conf})
	}

	for _, m := range emailRe.FindAllString(text, -1) {
		add("other", m, 0.95)
	}
	for _, m := range urlRe.FindAllString(text, -1) {
		add("other", m, 0.9)
	}
	for _, m := range ipv4Re.FindAllString(text, -1) {
		add("other", m, 0.8)
	}
	for _, m := range coordRe.FindAllString(text, -1) {
		add("location", m, 0.85)
	}
	for _, m := range hashtagRe.FindAllString(text, -1) {
		add("other", m, 0.7)
	}
	for _, m := range mentionRe.FindAllString(text, -1) {
		add("person", m, 0.6)
	}

	// Gazetteer: organizations.
	for _, m := range acronymRe.FindAllString(text, -1) {
		if orgGazetteer[m] {
			add("organization", m, 0.9)
		}
	}
	for name := range orgGazetteer {
		if strings.Contains(text, name) {
			add("organization", name, 0.95)
		}
	}
	// Gazetteer: countries (locations).
	lowerText := strings.ToLower(text)
	for _, c := range countryGazetteer {
		if strings.Contains(lowerText, strings.ToLower(c)) {
			add("location", c, 0.85)
		}
	}

	// Capitalized phrases as person candidates (low confidence).
	for _, m := range capPhraseRe.FindAllString(text, -1) {
		// Skip if it's a known country/org (already added) or sentence start heuristics.
		if countryGazetteerMatch(m) || orgGazetteer[m] {
			continue
		}
		// Only treat multi-word capitalized phrases as person candidates.
		if strings.Contains(m, " ") {
			add("person", m, 0.5)
		}
	}

	return out
}

func countryGazetteerMatch(s string) bool {
	for _, c := range countryGazetteer {
		if strings.EqualFold(c, s) {
			return true
		}
	}
	return false
}
