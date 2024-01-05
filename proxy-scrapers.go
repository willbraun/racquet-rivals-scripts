package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func scrapeWithProxy(targetURL string) string {
	proxyURL, err := url.Parse(os.Getenv("PROXY_URL"))
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		return ""
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		fmt.Println("Error making request:", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}

	return string(body)
}


// scrapeATP(draw DrawRecord) (slotSlice, map[string]string)
//   - proxyScrape the url to get the HTML
//   - use goquery to parse the HTML like before