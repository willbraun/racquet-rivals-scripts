package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Scraper interface {
	scrape(targetURL string) (string, error)
}

// Real scraper implementation

type RealScraper struct{}

func (r *RealScraper) scrape(targetURL string) (string, error) {
	printWithTimestamp("Visiting:", targetURL)

	proxyURL, err := url.Parse(os.Getenv("PROXY_URL"))
	if err != nil {
		log.Println(fmt.Sprintf("Error parsing proxy URL - %s:", targetURL), err)
		return "", fmt.Errorf("error parsing proxy URL: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		// Waiting for rendering can take a while so we set a longer timeout
		Timeout: 600 * time.Second,
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Bright Data header to wait for is-winner class to appear
	// Used for WTA draws to indicate that scores and winners have been rendered
	if strings.Contains(targetURL, "wtatennis.com") {
		req.Header.Set("x-unblock-expect", "{\"element\": \".match-table__tie-break\"}")
	}

	// Exponential backoff retry mechanism
	maxRetries := 5
	backoff := time.Second

	for i := range maxRetries {
		printWithTimestamp("Attempt:", i+1)
		resp, err := client.Do(req)
		if err != nil {
			log.Println(fmt.Sprintf("Error making request - %s:", targetURL), err)
			if i < maxRetries-1 {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return "", fmt.Errorf("error making request after %d retries: %w", maxRetries, err)
		}

		if resp.StatusCode != 200 {
			log.Println(fmt.Sprintf("HTTP error - %s:", targetURL), "Status Code:", resp.Status)
			resp.Body.Close() // Close the body before continuing
			if i < maxRetries-1 {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return "", fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(fmt.Sprintf("Error reading response body - %s:", targetURL), err)
			return "", fmt.Errorf("error reading response body: %w", err)
		}

		printWithTimestamp("Finished scraping:", targetURL)
		return string(body), nil
	}

	return "", fmt.Errorf("failed to scrape after %d retries", maxRetries)
}

// Scrapers for testing

func saveHTMLToFile(html, filename string) error {
	return os.WriteFile(filename, []byte(html), 0644)
}

func readHTMLFromFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type MockScraper struct{}

func (m *MockScraper) scrape(targetURL string) (string, error) {
	if strings.Contains(targetURL, "atptour.com") {
		html, err := readHTMLFromFile("scraped_pages/atp.html")
		if err != nil {
			log.Println("Error reading HTML from ATP file:", err)
			return "", fmt.Errorf("error reading ATP HTML file: %w", err)
		}
		return html, nil
	} else if strings.Contains(targetURL, "wtatennis.com") {
		html, err := readHTMLFromFile("scraped_pages/wta.html")
		if err != nil {
			log.Println("Error reading HTML from WTA file:", err)
			return "", fmt.Errorf("error reading WTA HTML file: %w", err)
		}
		return html, nil
	} else if strings.Contains(targetURL, "live-tennis.eu/en/wta-singles-draws") {
		html, err := readHTMLFromFile("scraped_pages/wta_live_tennis_eu.html")
		if err != nil {
			log.Println("Error reading HTML from WTA Live Tennis EU file:", err)
			return "", fmt.Errorf("error reading WTA Live Tennis EU HTML file: %w", err)
		}
		return html, nil
	}
	log.Println("Unknown URL:", targetURL)
	return "", fmt.Errorf("unknown URL: %s", targetURL)
}

type RealScraperSaveFile struct{}

func (s *RealScraperSaveFile) scrape(targetURL string) (string, error) {
	realScraper := &RealScraper{}
	html, err := realScraper.scrape(targetURL)
	if err != nil {
		return "", err
	}

	if strings.Contains(targetURL, "atptour.com") {
		err := saveHTMLToFile(html, "scraped_pages/atp.html")
		if err != nil {
			log.Println("Error saving ATP HTML to file:", err)
		}
	} else if strings.Contains(targetURL, "wtatennis.com") {
		err := saveHTMLToFile(html, "scraped_pages/wta.html")
		if err != nil {
			log.Println("Error saving WTA HTML to file:", err)
		}
	} else if strings.Contains(targetURL, "live-tennis.eu/en/wta-singles-draws") {
		err := saveHTMLToFile(html, "scraped_pages/wta_live_tennis_eu.html")
		if err != nil {
			log.Println("Error saving WTA Live Tennis EU HTML to file:", err)
		}
	}

	return html, nil
}

func getScraper(draw DrawRecord) Scraper {
	if os.Getenv("SAVE_HTML_TO_FILE") == "atp" && strings.Contains(draw.Url, "atptour.com") {
		return &RealScraperSaveFile{}
	} else if os.Getenv("SAVE_HTML_TO_FILE") == "wta" && strings.Contains(draw.Url, "wtatennis.com") {
		return &RealScraperSaveFile{}
	} else if os.Getenv("SAVE_HTML_TO_FILE") == "wta_live_tennis_eu" && strings.Contains(draw.Url, "live-tennis.eu/en/wta-singles-draws") {
		return &RealScraperSaveFile{}
	}
	return &MockScraper{}
}
