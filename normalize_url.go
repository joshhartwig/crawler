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
	parsedURL.Fragment = ""
	parsedURL.RawQuery = ""

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

	// the rawbaseurl can only be http or https
	if !strings.HasPrefix(rawBaseURL, "http://") && !strings.HasPrefix(rawBaseURL, "https://") {
		fmt.Printf("The rawbaseurl: %s does not start with http:// or https://", rawBaseURL)
		return nil, errors.New("couldn't parse base URL")
	}

	// remove any trailing /
	rawBaseURL = strings.TrimSuffix(rawBaseURL, "/")

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, r := range n.Attr {
				if r.Key == "href" {

					// if the url has '/'
					if strings.HasPrefix(r.Val, "/") {

						// if the url is only a '/'
						if r.Val == "/" {
							urls = append(urls, rawBaseURL) // just append the baseurl
							fmt.Printf("Found / only url skipping \n")
							return
						}

						// try and catch the duplication issue https://google.com/tags then appending /tags = google.com/tags/tags
						if strings.HasSuffix(rawBaseURL+"/", r.Val) {
							fmt.Printf("Found baseURL: %s with same suffix as URL: %s returning \n", rawBaseURL, r.Val)
							return
						}

						// for pass overlapping path test
						// TODO: pass the overlapping test

						fullURL := fmt.Sprintf("%s%s", rawBaseURL, r.Val)
						urls = append(urls, fullURL)
						fmt.Printf("Found relative url path: %s appending to baseurl: %s to form: %s \n", r.Val, rawBaseURL, fullURL)
						return
					} else {
						// if the url starts with http or https
						if strings.HasPrefix(r.Val, "http://") || strings.HasPrefix(r.Val, "https://") {
							fmt.Printf("Found absolute url: %s baseURL: %s \n", r.Val, rawBaseURL)
							urls = append(urls, r.Val)
							return
						} else {
							if strings.Contains(r.Val, "\\") {
								fmt.Printf("Found invalid url: %s baseURL: %s \n", r.Val, rawBaseURL)
								return
							}
							// if the baseurl has a suffix that is the same as the r.val /tags
							if !strings.HasSuffix(rawBaseURL, r.Val) {
								fullURL := fmt.Sprintf("%s/%s", rawBaseURL, r.Val)
								fmt.Printf("Found relative url: %s baseURL: %s \n", r.Val, rawBaseURL)
								urls = append(urls, fullURL)
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// return nil if we have nothing to return
	if len(urls) == 0 {
		return nil, nil
	}
	return urls, nil
}
