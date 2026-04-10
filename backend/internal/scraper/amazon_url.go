package scraper

import (
	"fmt"
	"net/url"
	"strings"
)

const affiliateTag = "prbox"

// RewriteAmazonURL strips any existing affiliate tag and injects tag=prbox.
// Works for amazon.co.uk, amazon.com, amzn.to short links, and URLs with
// existing tag= parameters anywhere in the query string.
func RewriteAmazonURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("empty URL")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}

	host := strings.ToLower(u.Host)
	if !strings.Contains(host, "amazon.") && !strings.Contains(host, "amzn.") {
		return "", fmt.Errorf("not an Amazon URL: %s", host)
	}

	q := u.Query()
	q.Del("tag")
	q.Set("tag", affiliateTag)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// IsAmazonURL returns true if the URL is an Amazon product page.
func IsAmazonURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	return strings.Contains(host, "amazon.") || strings.Contains(host, "amzn.")
}
