package main

import "testing"

func TestFindPriceCandidates(t *testing.T) {
	html := `
	<html><body>
		<span class="price">$39</span>
		<span class="price">Trial</span>
		<div id="product-price">14,99</div>
		<p>Some unrelated text paragraph with no price.</p>
		<span>not a price at all</span>
	</body></html>`

	doc := docFromString(html)
	candidates, err := FindPriceCandidates(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("expected at least one candidate, got none")
	}

	// Verify the known prices are found
	found3900 := false
	found1499 := false
	for _, c := range candidates {
		if c.Cents == 3900 {
			found3900 = true
		}
		if c.Cents == 1499 {
			found1499 = true
		}
	}

	if !found3900 {
		t.Errorf("expected candidate with 3900 cents ($39), not found in %v", candidates)
	}
	if !found1499 {
		t.Errorf("expected candidate with 1499 cents (14,99), not found in %v", candidates)
	}
}

func TestFindPriceCandidates_NoCandidates(t *testing.T) {
	html := `<html><body><p>No prices here at all.</p></body></html>`

	doc := docFromString(html)
	candidates, err := FindPriceCandidates(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(candidates) != 0 {
		t.Errorf("expected no candidates, got %d: %v", len(candidates), candidates)
	}
}
