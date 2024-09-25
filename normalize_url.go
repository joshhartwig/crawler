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
		fmt.Printf("url: %s does not start with http:// or https://", rawBaseURL)
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
							return
						}

						parsedURL, err := url.Parse(rawBaseURL)
						if err != nil {
							fmt.Println(err)
						}

						// Split the path of the base URL and href into slices
						basePaths := strings.Split(parsedURL.Path, "/") // Split base URL path (e.g., [tags])
						hrefPaths := strings.Split(r.Val, "/")          // Split href path (e.g., [tags, business])

						finalPath := "" // To construct the final URL path

						// Construct the combined path based on matching base URL paths and href
						for i := 0; i < len(hrefPaths); i++ {
							if i > len(basePaths)-1 { // If base URL path has fewer segments, append remaining href
								finalPath += hrefPaths[i] + "/"
								continue
							}
							// Only add path segments if they match
							if hrefPaths[i] == basePaths[i] {
								finalPath += hrefPaths[i] + "/"
							}
						}

						// Reconstruct the full URL by combining the scheme, host, and final path
						fullURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, finalPath)
						fullURL = strings.TrimSuffix(fullURL, "/") // Remove trailing slash from the final URL
						urls = append(urls, fullURL)
						fmt.Printf("relative url: %s => %s \n", r.Val, fullURL)
						return
					} else {
						// if the url starts with http or https
						if strings.HasPrefix(r.Val, "http://") || strings.HasPrefix(r.Val, "https://") {
							fmt.Printf("absolute url: %s \n", r.Val)
							urls = append(urls, r.Val)
							return
						} else {
							if strings.Contains(r.Val, "\\") {
								fmt.Printf("invalid url: %s \n", r.Val)
								return
							}
							// if the baseurl has a suffix that is the same as the r.val /tags
							if !strings.HasSuffix(rawBaseURL, r.Val) {
								fullURL := fmt.Sprintf("%s/%s", rawBaseURL, r.Val)
								fmt.Printf("relative url: %s => %s \n", r.Val, rawBaseURL)
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
