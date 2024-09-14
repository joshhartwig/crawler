package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func main() {
	args := os.Args

	if len(args) == 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(args) > 2 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	baseUrl := os.Args[1]
	fmt.Printf("starting crawl\n%s\n", baseUrl)

	// map to hold [url]count
	pages := map[string]int{}

	crawlPage(baseUrl, baseUrl, pages)

	fmt.Println(pages)
}

// fetch a URL and returns the html content as a string
func getHTML(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("no url")
	}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("error w/ http req %v", err)
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error creating client %v", err)
	}

	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error decoding resp body %v", err)
	}

	return string(bytes), nil
}

// recursively crawls a page getting urls from page and incrementing if we found existing
func crawlPage(baseURL, currentURL string, pages map[string]int) {
	// get the host name for both baseurl & current url and check if they are the same ex google.com
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("error parsing url", baseURL, err)
		return
	}

	parsedCurrentURL, err := url.Parse(currentURL)
	if err != nil {
		fmt.Println("error parsing url", currentURL, err)
		return
	}
	fmt.Printf("Is baseurl host: %s the same as currenturl host: %s \n", parsedBaseURL.Host, parsedCurrentURL.Host)
	// if we are not on the same hostname return, do not crawl the entire internet only urls from host
	if parsedBaseURL.Host != parsedCurrentURL.Host {
		fmt.Printf("Returning baseurl: %s and currenturl: %s are the same hosts \n", parsedBaseURL, parsedCurrentURL)
		return
	}

	// normalize the currentURL

	normalizedCurrentURL, err := normalizeURL(currentURL)
	if err != nil {
		fmt.Println("error normalizing url", err)
		return
	}
	fmt.Printf("normalizing current url: %s is normalized to: %s \n", currentURL, normalizedCurrentURL)

	// check if the normalized current URL is in the map, if so increment if not add it
	if _, ok := pages[normalizedCurrentURL]; !ok {
		pages[normalizedCurrentURL] = 1
		fmt.Printf("No entry in map for: %s setting value to 1 \n", normalizedCurrentURL)
	} else {
		pages[normalizedCurrentURL]++
		fmt.Printf("Entry in map found for: %s incrementing value \n", normalizedCurrentURL)
		return
	}

	// we have not crawled the page, so fetch html
	fmt.Printf("Fetching url: %s \n", currentURL)
	html, err := getHTML(currentURL)
	if err != nil {
		fmt.Println("error getting html", err)
		return
	}
	fmt.Printf("Fetched html with size: %b from %s\n", len([]byte(html)), currentURL)

	// get urls from html
	urls, err := getURLsFromHTML(html, currentURL)
	if err != nil {
		fmt.Println("error getting urls")
		return
	}

	// iterate through the urls and crawl
	for _, url := range urls {
		crawlPage(baseURL, url, pages)
	}
}
