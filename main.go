package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
)

// helper for sort
type kv struct {
	key string
	val int
}

type byVal []kv

func (a byVal) Len() int           { return len(a) }
func (a byVal) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byVal) Less(i, j int) bool { return a[i].val > a[j].val }

type byKey []kv

func (a byKey) Len() int           { return len(a) }
func (a byKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byKey) Less(i, j int) bool { return a[i].key < a[j].key }

type config struct {
	pages              map[string]int
	baseUrl            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func main() {

	args := os.Args
	maxConcurrency := 1
	maxPages := 10

	// if we only have 1 arg no site provided
	if len(args) == 1 {
		log.Println("no website provided")
		os.Exit(1)
	}

	// if we have more args then 4 we have an issue
	if len(args) > 4 {
		log.Println("too many arguments provided")
		os.Exit(1)
	}

	// fetch baseurl from args and parse
	baseUrl, err := url.Parse(os.Args[1])
	if err != nil {
		log.Println("error parsing url:", os.Args[1])
		os.Exit(1)
	}

	// fetch maxConcurrency and maxPages from our args
	if len(args) > 2 {
		fmt.Sscanf(args[2], "%d", &maxConcurrency)
	}

	if len(args) > 3 {
		fmt.Sscanf(args[3], "%d", &maxPages)
	}

	// stores our app configuration
	cfg := config{
		pages:              map[string]int{},
		baseUrl:            baseUrl,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}

	// Add waitgroup and kick off crawling
	cfg.wg.Add(1)
	go cfg.crawlPage(baseUrl.String())
	cfg.wg.Wait()

	// Print the final report
	printReport(cfg.pages, cfg.baseUrl.String())
}

// Fetch a URL and returns the html content as a string
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

// Recursively crawls a page getting urls from page and incrementing if we found existing
func (cfg *config) crawlPage(currentURL string) {
	defer cfg.wg.Done()
	cfg.concurrencyControl <- struct{}{} // send a struct into concurrencyControl

	if cfg.checkMapCount() > cfg.maxPages {
		<-cfg.concurrencyControl // release spot
		return
	}

	fmt.Printf("crawling: %s\n", currentURL)

	// Parse current url
	parsedCurrentURL, err := url.Parse(currentURL)
	if err != nil {
		<-cfg.concurrencyControl // release spot and return
		return
	}

	// if we are not on the same hostname return, do not crawl the entire internet only urls from host
	// ex if host is wagslane.dev vs cnn.com
	if cfg.baseUrl.Host != parsedCurrentURL.Host {
		normalizedParsedURL, _ := normalizeURL(parsedCurrentURL.String())
		cfg.addPageVisit(normalizedParsedURL)
		<-cfg.concurrencyControl // release spot
		return
	}

	// Normalize the currentURL
	normalizedCurrentURL, err := normalizeURL(currentURL)
	if err != nil {
		<-cfg.concurrencyControl // release spot
		return
	}

	// If this is the first time we have visited this page, get the html and start again
	if cfg.addPageVisit(normalizedCurrentURL) {
		html, err := getHTML(currentURL)
		if err != nil {
			<-cfg.concurrencyControl // release spot
			return
		}

		// get urls from html
		urls, err := getURLsFromHTML(html, currentURL)
		if err != nil {
			<-cfg.concurrencyControl // release spot
			return
		}

		// iterate through the urls and crawl
		for _, url := range urls {
			cfg.wg.Add(1)
			go cfg.crawlPage(url)
		}
	}

	<-cfg.concurrencyControl

}

// Check if the normalized url is in our map, if not add it, if so increment the count
func (c *config) addPageVisit(normalizedUrl string) (isFirst bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// if we did not find the entry in our map, this is the first time we have seen this page
	if _, ok := c.pages[normalizedUrl]; !ok {
		isFirst = true
		c.pages[normalizedUrl] = 1
		return isFirst
	}

	isFirst = false
	c.pages[normalizedUrl]++
	return isFirst
}

// Retuns the size of our map
func (c *config) checkMapCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	mapSize := len(c.pages)
	return mapSize
}

// Prints our final report - calls sorting method for pages
func printReport(pages map[string]int, baseURL string) {
	fmt.Println("=============================")
	fmt.Printf("REPORT for %s\n", baseURL)
	fmt.Println("=============================")

	results := sortMap(pages)
	for _, kv := range results {
		fmt.Printf("Found %d internal links to %s \n", kv.val, kv.key)
	}
}

// Sorts map and returns slice of key val structs
func sortMap(p map[string]int) []kv {
	ss := []kv{}
	for k, v := range p {
		ss = append(ss, kv{k, v})
	}

	sort.Sort(byVal(ss))
	//sort.Stable(byKey(ss))

	return ss
}
