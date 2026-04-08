// Package main provides functionality to fetch traffic values.
package main

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// FetchValues fetches available traffic URLs from the API.
func FetchValues() ([]string, error) {
	resp, err := http.Get("https://lolicp.com/api/wasted_traffic_plus/")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html failed: %w", err)
	}

	var values []string
	doc.Find("select#select option").Each(func(_ int, selection *goquery.Selection) {
		value, exists := selection.Attr("value")
		if !exists || value == "" {
			return
		}
		values = append(values, value)
	})

	if len(values) == 0 {
		return nil, fmt.Errorf("no values found")
	}

	return values, nil
}
