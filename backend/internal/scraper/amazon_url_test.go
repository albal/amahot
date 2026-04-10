package scraper

import (
	"strings"
	"testing"
)

func TestRewriteAmazonURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string // substring that must appear in result
		wantErr bool
	}{
		{
			name:  "adds tag to plain ASIN URL",
			input: "https://www.amazon.co.uk/dp/B09XS7JWHH",
			want:  "tag=prbox",
		},
		{
			name:  "replaces existing tag",
			input: "https://www.amazon.co.uk/dp/B09XS7JWHH?tag=hotukdeals-21",
			want:  "tag=prbox",
		},
		{
			name:  "does not keep old tag",
			input: "https://www.amazon.co.uk/dp/B09XS7JWHH?tag=someone-21",
			want:  "tag=prbox",
		},
		{
			name:  "preserves other query params",
			input: "https://www.amazon.co.uk/dp/B09XS7JWHH?th=1&psc=1",
			want:  "th=1",
		},
		{
			name:  "amazon.com URL",
			input: "https://www.amazon.com/dp/B09XS7JWHH",
			want:  "tag=prbox",
		},
		{
			name:  "amzn.to short link",
			input: "https://amzn.to/3xYZabc",
			want:  "tag=prbox",
		},
		{
			name:  "strips old tag from middle of query string",
			input: "https://www.amazon.co.uk/dp/B0?foo=1&tag=old-22&bar=2",
			want:  "tag=prbox",
		},
		{
			name:    "empty URL returns error",
			input:   "",
			wantErr: true,
		},
		{
			name:    "non-Amazon URL returns error",
			input:   "https://www.hotukdeals.com/deals/some-deal-123",
			wantErr: true,
		},
		{
			name:    "invalid URL returns error",
			input:   "://not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RewriteAmazonURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("RewriteAmazonURL(%q) = %q, want it to contain %q", tt.input, got, tt.want)
			}
			// Old tags must be gone
			if strings.Contains(tt.input, "tag=") && strings.Count(got, "tag=") != 1 {
				t.Errorf("result has multiple tag= params: %q", got)
			}
		})
	}
}

func TestRewriteAmazonURL_TagValue(t *testing.T) {
	// Confirm the exact tag value is "prbox" with no suffix
	got, err := RewriteAmazonURL("https://www.amazon.co.uk/dp/B09XS7JWHH")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "tag=prbox") {
		t.Errorf("expected tag=prbox in %q", got)
	}
	// Must not have tag=prbox-21 or tag=prbox-20
	if strings.Contains(got, "tag=prbox-") {
		t.Errorf("tag should be exactly 'prbox', got: %q", got)
	}
}

func TestIsAmazonURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://www.amazon.co.uk/dp/B09XS7JWHH", true},
		{"https://www.amazon.com/dp/B09XS7JWHH", true},
		{"https://amzn.to/3xYZabc", true},
		{"https://amzn.eu/d/someproduct", true},
		{"https://www.hotukdeals.com/deals/some-deal-123", false},
		{"https://www.google.com", false},
		{"not-a-url", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := IsAmazonURL(tt.url)
			if got != tt.want {
				t.Errorf("IsAmazonURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
