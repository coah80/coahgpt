package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Result struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebSearch performs a DuckDuckGo search and returns results
func WebSearch(query string, maxResults int) ([]Result, error) {
	if maxResults <= 0 {
		maxResults = 5
	}

	// Use DuckDuckGo HTML lite for scraping
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; coahGPT/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseDDGResults(string(body), maxResults), nil
}

// parseDDGResults extracts results from DuckDuckGo HTML
func parseDDGResults(html string, max int) []Result {
	var results []Result

	// Simple extraction of result snippets from DDG HTML
	parts := strings.Split(html, "result__a")
	for i, part := range parts {
		if i == 0 || len(results) >= max {
			continue
		}

		// extract href
		hrefStart := strings.Index(part, "href=\"")
		if hrefStart == -1 {
			continue
		}
		hrefStart += 6
		hrefEnd := strings.Index(part[hrefStart:], "\"")
		if hrefEnd == -1 {
			continue
		}
		rawURL := part[hrefStart : hrefStart+hrefEnd]

		// decode DDG redirect URL
		actualURL := rawURL
		if strings.Contains(rawURL, "uddg=") {
			if u, err := url.Parse(rawURL); err == nil {
				if uddg := u.Query().Get("uddg"); uddg != "" {
					actualURL = uddg
				}
			}
		}

		// extract title (text between > and </a>)
		titleStart := strings.Index(part[hrefEnd:], ">")
		if titleStart == -1 {
			continue
		}
		titleEnd := strings.Index(part[hrefEnd+titleStart:], "</a>")
		if titleEnd == -1 {
			continue
		}
		title := stripTags(part[hrefEnd+titleStart+1 : hrefEnd+titleStart+titleEnd])

		// extract snippet
		snippet := ""
		snippetStart := strings.Index(part, "result__snippet")
		if snippetStart != -1 {
			snipStart := strings.Index(part[snippetStart:], ">")
			if snipStart != -1 {
				snipEnd := strings.Index(part[snippetStart+snipStart:], "</")
				if snipEnd != -1 {
					snippet = stripTags(part[snippetStart+snipStart+1 : snippetStart+snipStart+snipEnd])
				}
			}
		}

		if title != "" && actualURL != "" {
			results = append(results, Result{
				Title:   strings.TrimSpace(title),
				URL:     actualURL,
				Snippet: strings.TrimSpace(snippet),
			})
		}
	}

	return results
}

func stripTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// FormatResults formats search results as context for the LLM
func FormatResults(results []Result) string {
	if len(results) == 0 {
		return "No search results found."
	}

	var sb strings.Builder
	sb.WriteString("Web search results:\n\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", r.URL))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("   %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// FormatResultsJSON returns results as JSON
func FormatResultsJSON(results []Result) string {
	data, _ := json.Marshal(results)
	return string(data)
}
