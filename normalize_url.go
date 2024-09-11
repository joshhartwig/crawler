package main

import (
	"errors"
	"fmt"
	"strings"

	"net/url"

	"golang.org/x/net/html"
)

// returns a whole url if it comes across a relative path url like an href to /home => https://site.com/home
func normalizeURL(inputUrl string) (string, error) {
	inputUrl = strings.ToLower(inputUrl)
	parsedURL, err := url.Parse(inputUrl)
	if err != nil {
		return "", errors.New("couldn't parse URL")
	}

	parsedURL.Path = strings.TrimRight(parsedURL.Path, "/")

	return fmt.Sprintf("%s%s", parsedURL.Hostname(), parsedURL.Path), nil
}

// returns a slice of unnormalized urls from raw html
func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {

	urls := []string{}
	body := strings.NewReader(htmlBody)
	doc, err := html.Parse(body)
	if err != nil {
		return urls, fmt.Errorf("couldn't parse base URL %v", err)
	}

	// if the rawBaseUrl is a valid url proceed
	if strings.HasPrefix(rawBaseURL, "http://") || strings.HasPrefix(rawBaseURL, "https://") {
		var f func(*html.Node)
		f = func(n *html.Node) {
			// if we find an 'a' tag
			if n.Type == html.ElementNode && n.Data == "a" {

				// iterate through attribs
				for _, r := range n.Attr {

					// if we contain a "\" its not valid url
					if strings.Contains(r.Val, "\\") {
						break
					}

					// if the first char is not a / & rawbaseurl does not end w/ slash
					if strings.Index(r.Val, "/") != 0 && !strings.HasSuffix(rawBaseURL, "/") {

						// if its a full url append it and return
						if strings.HasPrefix(r.Val, "http://") || strings.HasPrefix(r.Val, "https://") {
							urls = append(urls, r.Val)
						} else {
							urls = append(urls, fmt.Sprintf("%s/%s", rawBaseURL, r.Val))
							break
						}

					}

					// if the 1st char is a / then r.val is a path then append to baseurl and add to slice
					if strings.Index(r.Val, "/") == 0 {
						urls = append(urls, fmt.Sprintf("%s%s", rawBaseURL, r.Val))
						break
					}

				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)
	} else {
		return nil, fmt.Errorf("couldn't parse base URL")
	}

	// return nil if we have nothing to return
	if len(urls) == 0 {
		return nil, nil
	}
	return urls, nil
}
