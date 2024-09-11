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

	// if we are not on the same hostname return, do not crawl the entire internet only urls from host
	if parsedBaseURL.Hostname() != parsedCurrentURL.Hostname() {
		fmt.Printf("returning as the baseurl %s and currenturl %s are the same hosts", parsedBaseURL, parsedCurrentURL)
		return
	}

	// normalize the currentURL
	fmt.Printf("normalizing current url: %s\n", currentURL)
	normalizedCurrentURL, err := normalizeURL(currentURL)
	if err != nil {
		fmt.Println("error normalizing url", err)
		return
	}

	// check if the normalized current URL is in the map, if so increment if not add it
	if _, ok := pages[normalizedCurrentURL]; !ok {
		pages[normalizedCurrentURL] = 1
		fmt.Printf("No entry in map for: %s setting value to 1\n", normalizedCurrentURL)
	} else {
		pages[normalizedCurrentURL]++
		fmt.Printf("Entry in map found for: %s incrementing value \n", normalizedCurrentURL)
		return // we already crawled this page
	}

	// we have not crawled the page, so fetch html
	fmt.Printf("fetching %s\n", currentURL)
	html, err := getHTML(currentURL)
	if err != nil {
		fmt.Println("error getting html", err)
		return
	}
	fmt.Printf("fetched html with size: %db from %s\n", len([]byte(html)), currentURL)

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

/*
1. crawl gets called crawl("https://wagnerlane.com","https://wagnerlane.com", pages)
2. check if base & current = same host if not stop
3. normalize the rawcurrent
4. check if the normalized url is in pages if not add it, if so increment
5. fetch html for normalizedrawcurrent
6. get the urls from the html
7. iter over urls and call crawl on those urls

1.

*/
