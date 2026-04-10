package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/albal/amahot/backend/internal/models"
	"github.com/albal/amahot/backend/internal/repository"
)

// hotukdeals uses Vue3 components; deal data lives in JSON inside data-vue3 attributes.
// The ThreadMainListItemNormalizer component holds the full thread payload.

type threadPayload struct {
	Name  string `json:"name"`
	Props struct {
		Thread threadData `json:"thread"`
	} `json:"props"`
}

type threadData struct {
	ThreadID    string  `json:"threadId"`
	TitleSlug   string  `json:"titleSlug"`
	Title       string  `json:"title"`
	Temperature float64 `json:"temperature"`
	Price       float64 `json:"price"`
	Link        string  `json:"link"`
	LinkHost    string  `json:"linkHost"`
	IsExpired   bool    `json:"isExpired"`
	Merchant    struct {
		MerchantName string `json:"merchantName"`
	} `json:"merchant"`
	MainImage struct {
		Path string `json:"path"`
		Name string `json:"name"`
		Ext  string `json:"ext"`
	} `json:"mainImage"`
	MainGroup struct {
		ThreadGroupName string `json:"threadGroupName"`
	} `json:"mainGroup"`
}

const (
	minTemperature = 100.0
	maxPages       = 3
	pageDelaySecs  = 3
)

type Scraper struct {
	baseURL   string
	userAgent string
	dealRepo  *repository.DealRepo
	client    *http.Client
}

func New(baseURL, userAgent string, dealRepo *repository.DealRepo) *Scraper {
	return &Scraper{
		baseURL:   baseURL,
		userAgent: userAgent,
		dealRepo:  dealRepo,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Scraper) Run(ctx context.Context) {
	log.Println("Scraper: starting run")
	total := 0

	for page := 1; page <= maxPages; page++ {
		if page > 1 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(pageDelaySecs * time.Second):
			}
		}

		pageURL := fmt.Sprintf("%s&page=%d", s.baseURL, page)
		count, err := s.scrapePage(ctx, pageURL)
		if err != nil {
			log.Printf("Scraper: page %d error: %v", page, err)
			break
		}
		if count == 0 {
			log.Printf("Scraper: page %d returned 0 deals, stopping", page)
			break
		}
		total += count
		log.Printf("Scraper: page %d — %d deals processed", page, count)
	}

	log.Printf("Scraper: run complete, %d deals upserted", total)
}

func (s *Scraper) scrapePage(ctx context.Context, pageURL string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.9")
	// Do NOT set Accept-Encoding manually — Go's http.Transport handles gzip
	// negotiation and transparent decompression automatically. Setting it
	// manually causes double-decompression issues.
	req.Header.Set("Referer", "https://www.google.com/")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetch %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, pageURL)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("parse HTML: %w", err)
	}

	count := 0
	doc.Find("div[data-vue3*=\"ThreadMainListItemNormalizer\"]").Each(func(_ int, sel *goquery.Selection) {
		raw, exists := sel.Attr("data-vue3")
		if !exists {
			return
		}

		var payload threadPayload
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return
		}

		thread := payload.Props.Thread
		if thread.ThreadID == "" || thread.Title == "" {
			return
		}
		if thread.Temperature < minTemperature {
			return
		}
		if thread.IsExpired {
			return
		}
		if !isAmazonHost(thread.LinkHost) {
			return
		}

		deal := buildDeal(thread)
		if err := s.dealRepo.Upsert(ctx, deal); err != nil {
			log.Printf("Scraper: upsert deal %s: %v", deal.ExternalID, err)
			return
		}
		count++
	})

	return count, nil
}

func buildDeal(t threadData) models.Deal {
	// Construct deal URL: if the link field is set use it (rare), otherwise
	// use the hotukdeals deal page URL which redirects to Amazon when the user
	// clicks "Get Deal" there.
	dealURL := t.Link
	if dealURL == "" {
		dealURL = fmt.Sprintf(
			"https://www.hotukdeals.com/deals/%s-%s",
			t.TitleSlug,
			t.ThreadID,
		)
	}

	// Rewrite Amazon URLs with affiliate tag; hotukdeals URLs are kept as-is.
	if IsAmazonURL(dealURL) {
		if rewritten, err := RewriteAmazonURL(dealURL); err == nil {
			dealURL = rewritten
		}
	}

	// Image URL from hotukdeals CDN
	imageURL := ""
	if t.MainImage.Path != "" && t.MainImage.Name != "" {
		ext := t.MainImage.Ext
		if ext == "" {
			ext = "jpg"
		}
		imageURL = fmt.Sprintf(
			"https://static.hotukdeals.com/%s/%s.%s",
			t.MainImage.Path,
			t.MainImage.Name,
			ext,
		)
	}

	price := ""
	if t.Price > 0 {
		price = fmt.Sprintf("£%.2f", t.Price)
	}

	merchant := t.Merchant.MerchantName
	if merchant == "" {
		merchant = "Amazon"
	}

	temp := int(math.Round(t.Temperature))

	return models.Deal{
		ExternalID:  t.ThreadID,
		Title:       t.Title,
		Price:       price,
		ImageURL:    imageURL,
		DealURL:     dealURL,
		Merchant:    merchant,
		Temperature: temp,
		Category:    t.MainGroup.ThreadGroupName,
	}
}

func isAmazonHost(host string) bool {
	h := strings.ToLower(host)
	return strings.Contains(h, "amazon.")
}
