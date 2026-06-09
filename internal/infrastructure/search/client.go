package search

import (
	"context"
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

type Client struct {
	http *http.Client
	sxng string // SearXNG URL if configured
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) Search(ctx context.Context, query string) ([]Result, error) {
	// Try SearXNG first if configured
	if c.sxng != "" {
		if results, err := c.searchSearXNG(ctx, query); err == nil && len(results) > 0 {
			return results, nil
		}
	}
	// Fallback to DuckDuckGo lite
	return c.searchDuckDuckGo(ctx, query)
}

func (c *Client) searchDuckDuckGo(ctx context.Context, query string) ([]Result, error) {
	u := fmt.Sprintf("https://lite.duckduckgo.com/lite/?q=%s", url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseDuckDuckGoResults(string(body)), nil
}

func parseDuckDuckGoResults(html string) []Result {
	var results []Result
	lines := strings.Split(html, "\n")
	for i, line := range lines {
		// DuckDuckGo lite uses <a href="..."> for result links
		if strings.HasPrefix(line, "<a href=\"") && strings.Contains(line, "</a>") {
			start := strings.Index(line, "\"")
			end := strings.Index(line[start+1:], "\"")
			if start < 0 || end < 0 {
				continue
			}
			link := line[start+1 : start+1+end]
			// Extract text between <a> and </a>
			aStart := strings.Index(line, ">") + 1
			aEnd := strings.LastIndex(line, "</a>")
			if aStart < 0 || aEnd < 0 || aStart >= aEnd {
				continue
			}
			title := strings.TrimSpace(line[aStart:aEnd])
			// Get snippet from next td
			snippet := ""
			if i+2 < len(lines) {
				s := strings.TrimSpace(lines[i+2])
				if s != "" {
					snippet = s
				}
			}
			if title != "" && !strings.HasPrefix(link, "/") {
				results = append(results, Result{Title: title, URL: link, Snippet: snippet})
			}
		}
	}
	if len(results) > 10 {
		results = results[:10]
	}
	return results
}

func (c *Client) searchSearXNG(ctx context.Context, query string) ([]Result, error) {
	u := fmt.Sprintf("%s/search?q=%s&format=json", c.sxng, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Simple JSON parse
	text := string(body)
	var results []Result
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, `"title":`) {
			results = append(results, Result{
				Title:   extractJSONField(line, "title"),
				URL:     extractJSONField(line, "url"),
				Snippet: extractJSONField(line, "content"),
			})
		}
	}
	return results, nil
}

func extractJSONField(line, field string) string {
	pattern := fmt.Sprintf(`"%s":"`, field)
	start := strings.Index(line, pattern)
	if start < 0 {
		return ""
	}
	start += len(pattern)
	end := strings.Index(line[start:], `"`)
	if end < 0 {
		return ""
	}
	return line[start : start+end]
}
