package scraper

import (
	"strings"
	"testing"
)

func TestIsAmazonHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"www.amazon.co.uk", true},
		{"www.amazon.com", true},
		{"amazon.de", true},
		{"www.hotukdeals.com", false},
		{"www.ebay.co.uk", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := isAmazonHost(tt.host); got != tt.want {
				t.Errorf("isAmazonHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestBuildDeal_URLFallback(t *testing.T) {
	// When link is empty, deal URL should be the hotukdeals deal page.
	td := threadData{
		ThreadID:    "1234567",
		TitleSlug:   "some-cool-product",
		Title:       "Some Cool Product",
		Temperature: 142.7,
		LinkHost:    "www.amazon.co.uk",
		Merchant:    struct{ MerchantName string `json:"merchantName"` }{"Amazon"},
	}
	deal := buildDeal(td)

	want := "https://www.hotukdeals.com/deals/some-cool-product-1234567"
	if deal.DealURL != want {
		t.Errorf("DealURL = %q, want %q", deal.DealURL, want)
	}
}

func TestBuildDeal_AmazonLinkRewritten(t *testing.T) {
	// When link is a direct Amazon URL it must get the affiliate tag.
	td := threadData{
		ThreadID:    "9999",
		TitleSlug:   "test-product",
		Title:       "Test Product",
		Temperature: 200,
		Link:        "https://www.amazon.co.uk/dp/B0TESTTEST",
		LinkHost:    "www.amazon.co.uk",
		Merchant:    struct{ MerchantName string `json:"merchantName"` }{"Amazon"},
	}
	deal := buildDeal(td)

	if !strings.Contains(deal.DealURL, "tag=prbox") {
		t.Errorf("DealURL missing affiliate tag: %q", deal.DealURL)
	}
	if strings.Contains(deal.DealURL, "hotukdeals.com") {
		t.Errorf("DealURL should be Amazon URL, not hotukdeals: %q", deal.DealURL)
	}
}

func TestBuildDeal_AmazonLinkExistingTagReplaced(t *testing.T) {
	td := threadData{
		ThreadID:  "8888",
		TitleSlug: "product",
		Title:     "Product",
		Link:      "https://www.amazon.co.uk/dp/B0TEST?tag=hotukdeals-21&th=1",
		LinkHost:  "www.amazon.co.uk",
	}
	deal := buildDeal(td)

	if strings.Contains(deal.DealURL, "hotukdeals-21") {
		t.Errorf("old affiliate tag should be stripped: %q", deal.DealURL)
	}
	if !strings.Contains(deal.DealURL, "tag=prbox") {
		t.Errorf("prbox tag missing: %q", deal.DealURL)
	}
}

func TestBuildDeal_PriceFormatting(t *testing.T) {
	tests := []struct {
		price     float64
		wantPrice string
	}{
		{8.86, "£8.86"},
		{22.50, "£22.50"},
		{0, ""},
		{1.99, "£1.99"},
		{100, "£100.00"},
	}
	for _, tt := range tests {
		td := threadData{ThreadID: "1", TitleSlug: "p", Title: "P", Price: tt.price}
		deal := buildDeal(td)
		if deal.Price != tt.wantPrice {
			t.Errorf("Price(%v) = %q, want %q", tt.price, deal.Price, tt.wantPrice)
		}
	}
}

func TestBuildDeal_MerchantDefault(t *testing.T) {
	// Empty merchant name should fall back to "Amazon".
	td := threadData{ThreadID: "1", TitleSlug: "p", Title: "P"}
	deal := buildDeal(td)
	if deal.Merchant != "Amazon" {
		t.Errorf("Merchant = %q, want %q", deal.Merchant, "Amazon")
	}
}

func TestBuildDeal_MerchantPreserved(t *testing.T) {
	td := threadData{
		ThreadID:  "1",
		TitleSlug: "p",
		Title:     "P",
		Merchant:  struct{ MerchantName string `json:"merchantName"` }{"Amazon Warehouse"},
	}
	deal := buildDeal(td)
	if deal.Merchant != "Amazon Warehouse" {
		t.Errorf("Merchant = %q, want %q", deal.Merchant, "Amazon Warehouse")
	}
}

func TestBuildDeal_TemperatureRounded(t *testing.T) {
	tests := []struct {
		temp float64
		want int
	}{
		{102.43, 102},
		{102.5, 103},
		{199.9, 200},
		{500.0, 500},
	}
	for _, tt := range tests {
		td := threadData{ThreadID: "1", TitleSlug: "p", Title: "P", Temperature: tt.temp}
		deal := buildDeal(td)
		if deal.Temperature != tt.want {
			t.Errorf("Temperature(%v) = %d, want %d", tt.temp, deal.Temperature, tt.want)
		}
	}
}

func TestBuildDeal_ImageURL(t *testing.T) {
	td := threadData{
		ThreadID:  "1",
		TitleSlug: "p",
		Title:     "P",
	}
	td.MainImage.Path = "threads/raw/ojbvc"
	td.MainImage.Name = "4863744_1"
	td.MainImage.Ext = "jpg"

	deal := buildDeal(td)
	want := "https://static.hotukdeals.com/threads/raw/ojbvc/4863744_1.jpg"
	if deal.ImageURL != want {
		t.Errorf("ImageURL = %q, want %q", deal.ImageURL, want)
	}
}

func TestBuildDeal_ImageURLEmptyWhenMissing(t *testing.T) {
	td := threadData{ThreadID: "1", TitleSlug: "p", Title: "P"}
	deal := buildDeal(td)
	if deal.ImageURL != "" {
		t.Errorf("ImageURL should be empty when no image data, got %q", deal.ImageURL)
	}
}

func TestBuildDeal_ExternalID(t *testing.T) {
	td := threadData{ThreadID: "4863744", TitleSlug: "p", Title: "P"}
	deal := buildDeal(td)
	if deal.ExternalID != "4863744" {
		t.Errorf("ExternalID = %q, want %q", deal.ExternalID, "4863744")
	}
}

func TestExtractAmazonProductURL(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "bare URL in description",
			html: `Check out this deal: https://www.amazon.co.uk/dp/B09XS7JWHH great price`,
			want: "https://www.amazon.co.uk/dp/B09XS7JWHH",
		},
		{
			name: "URL with query params",
			html: `<a href="https://www.amazon.co.uk/dp/B0TESTTEST?th=1&amp;psc=1">link</a>`,
			want: "https://www.amazon.co.uk/dp/B0TESTTEST?th=1&amp;psc=1",
		},
		{
			name: "URL with product slug prefix",
			html: `https://www.amazon.co.uk/Sony-Headphones-WH1000XM5/dp/B09XS7JWHH?tag=old-21`,
			want: "https://www.amazon.co.uk/Sony-Headphones-WH1000XM5/dp/B09XS7JWHH?tag=old-21",
		},
		{
			name: "no Amazon URL",
			html: `<p>Great deal on hotukdeals!</p>`,
			want: "",
		},
		{
			name: "empty string",
			html: "",
			want: "",
		},
		{
			name: "URL inside anchor tag href",
			html: `Buy here <a href="https://www.amazon.co.uk/dp/B001234567">Amazon</a>`,
			want: "https://www.amazon.co.uk/dp/B001234567",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractAmazonProductURL(tt.html)
			if got != tt.want {
				t.Errorf("extractAmazonProductURL(%q) = %q, want %q", tt.html, got, tt.want)
			}
		})
	}
}
