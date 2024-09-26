package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/google/uuid"
)

type config struct {
	pages              map[string]int
	baseUrl            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
}

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

	baseUrl, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Println("error parsing url:", os.Args[1])
	}

	fmt.Printf("starting crawl\n%s\n", baseUrl.String())

	cfg := config{
		pages:              map[string]int{},
		baseUrl:            baseUrl,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, 5),
		wg:                 &sync.WaitGroup{},
	}

	cfg.wg.Add(1)
	go cfg.crawlPage(baseUrl.String())
	cfg.wg.Wait()

	prettyPrintMap(cfg.pages)
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
	uuid := uuid.New()
	fmt.Printf("crawling: %s in routine: %s\n", currentURL, uuid.String())

	defer cfg.wg.Done()
	cfg.concurrencyControl <- struct{}{} // aquire a spot

	parsedCurrentURL, err := url.Parse(currentURL)
	if err != nil {
		<-cfg.concurrencyControl // release spot
		return
	}

	// if we are not on the same hostname return, do not crawl the entire internet only urls from host
	// ex if host is wagslane.dev vs cnn.com
	if cfg.baseUrl.Host != parsedCurrentURL.Host {
		fmt.Println("we are not the same host baseurlhost: ", cfg.baseUrl.Host, " parsedcurrneturlhost: ", parsedCurrentURL.Host)
		normalizedParsedURL, _ := normalizeURL(parsedCurrentURL.String())
		cfg.addPageVisit(normalizedParsedURL)
		<-cfg.concurrencyControl // release spot
		return
	}

	// normalize the currentURL
	normalizedCurrentURL, err := normalizeURL(currentURL)
	if err != nil {
		fmt.Println("error normalizing url", err)
		<-cfg.concurrencyControl // release spot
		return
	}

	// if this is the first time we have visited this page, get the html and start again
	if cfg.addPageVisit(normalizedCurrentURL) {
		// we have not crawled the page, so fetch html
		html, err := getHTML(currentURL)
		fmt.Printf("fetching html for: %s in routine: %s \n", currentURL, uuid.String())
		if err != nil {
			fmt.Println("error getting html", err)
			<-cfg.concurrencyControl // release spot
			return
		}

		// get urls from html
		urls, err := getURLsFromHTML(html, currentURL)
		if err != nil {
			fmt.Println("error getting urls")
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

func (c *config) addPageVisit(normalizedUrl string) (isFirst bool) {
	c.mu.Lock()
	// if we did not find the entry in our map, this is the first time we have seen this page
	if _, ok := c.pages[normalizedUrl]; !ok {
		isFirst = true
		fmt.Println("new entry (adding to map) :", normalizedUrl)
		c.pages[normalizedUrl] = 1
		c.mu.Unlock()
		return isFirst
	}

	isFirst = false
	fmt.Println("skipping... already seen: ", normalizedUrl)
	c.pages[normalizedUrl]++
	c.mu.Unlock()
	return isFirst
}

func prettyPrintMap(m map[string]int) {
	fmt.Println()
	fmt.Println()
	fmt.Printf("printing results of crawl found %d entries \n", len(m))
	fmt.Println("-----------------------------------------------")
	for k, v := range m {
		fmt.Printf("%s - %d entries \n", k, v)
	}
}
