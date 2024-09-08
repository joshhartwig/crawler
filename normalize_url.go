package main

import (
	"fmt"
	"strings"

	"net/url"

	"golang.org/x/net/html"
)

// returns a whole url if it comes across a relative path url like an href to /home => https://site.com/home
func normalizeURL(inputUrl string) (string, error) {
	// lower case everything
	inputUrl = strings.ToLower(inputUrl)

	parsed, err := url.Parse(inputUrl)
	if err != nil {
		return "", fmt.Errorf("couldn't parse URL: %w", err)
	}

	path := parsed.Path
	path = strings.TrimSuffix(path, "/")

	return fmt.Sprintf("%s%s", parsed.Hostname(), path), nil
}

// returns a slice of unnormalized urls from raw html
func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	urls := []string{}
	body := strings.NewReader(htmlBody)
	doc, err := html.Parse(body)
	if err != nil {
		return []string{}, fmt.Errorf("error %v", err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, r := range n.Attr {
				// if the 1st char is a / then r.val is a path not full url
				if strings.Index(r.Val, "/") == 0 {
					urls = append(urls, fmt.Sprintf("%s%s", rawBaseURL, r.Val))
					break
				}
				// do not append malformed urls
				if strings.HasPrefix(r.Val, "http://") || strings.HasPrefix(r.Val, "https://") {
					urls = append(urls, r.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return urls, nil
}
