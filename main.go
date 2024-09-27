package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
)

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

	if len(args) == 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(args) > 4 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	baseUrl, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Println("error parsing url:", os.Args[1])
	}

	if len(args) > 2 {
		fmt.Sscanf(args[2], "%d", &maxConcurrency)
	}

	if len(args) > 3 {
		fmt.Sscanf(args[3], "%d", &maxPages)
	}

	cfg := config{
		pages:              map[string]int{},
		baseUrl:            baseUrl,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}

	cfg.wg.Add(1)
	go cfg.crawlPage(baseUrl.String())
	cfg.wg.Wait()

	printReport(cfg.pages, cfg.baseUrl.String())
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
func (cfg *config) crawlPage(currentURL string) {
	defer cfg.wg.Done()
	cfg.concurrencyControl <- struct{}{} // aquire a spot

	if cfg.checkMapCount() > cfg.maxPages {
		<-cfg.concurrencyControl // release spot
		return
	}

	fmt.Printf("crawling: %s\n", currentURL)

	parsedCurrentURL, err := url.Parse(currentURL)
	if err != nil {
		<-cfg.concurrencyControl // release spot
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

	// normalize the currentURL
	normalizedCurrentURL, err := normalizeURL(currentURL)
	if err != nil {
		<-cfg.concurrencyControl // release spot
		return
	}

	// if this is the first time we have visited this page, get the html and start again
	if cfg.addPageVisit(normalizedCurrentURL) {
		// we have not crawled the page, so fetch html
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

// check if the normalized url is in our map, if not add it, if so increment the count
func (c *config) addPageVisit(normalizedUrl string) (isFirst bool) {
	c.mu.Lock()
	// if we did not find the entry in our map, this is the first time we have seen this page
	if _, ok := c.pages[normalizedUrl]; !ok {
		isFirst = true
		c.pages[normalizedUrl] = 1
		c.mu.Unlock()
		return isFirst
	}

	isFirst = false
	c.pages[normalizedUrl]++
	c.mu.Unlock()
	return isFirst
}

// retuns the size of our map
func (c *config) checkMapCount() int {
	c.mu.Lock()
	mapSize := len(c.pages)
	c.mu.Unlock()
	return mapSize
}

func printReport(pages map[string]int, baseURL string) {
	fmt.Println("=============================")
	fmt.Printf("REPORT for %s\n", baseURL)
	fmt.Println("=============================")

	results := sortMap(pages)
	for _, kv := range results {
		fmt.Printf("Found %d internal links to %s \n", kv.val, kv.key)
	}
}

// helper for sort
type kv struct {
	key string
	val int
}

// sorts map and returns slice of key val structs
func sortMap(p map[string]int) []kv {
	ss := []kv{}
	for k, v := range p {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].val > ss[j].val
	})

	return ss
}
