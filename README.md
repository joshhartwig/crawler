# Go Web Crawler

A simple, concurrent web crawler written in Go that recursively crawls a website to collect and count internal links. The crawler starts from a given base URL and captures all the internal links up to a specified limit. Built for the kicks, good way to learn concurrency in Go.

## Features

- Recursively crawls internal links of a website
- Concurrency control for efficient crawling
- Handles relative and absolute URLs
- Generates a report of all the visited links and their frequencies
- Allows setting limits on maximum pages and concurrency level

## Requirements

- Go 1.19+ (or any version that supports Go modules)

## Installation

1. Clone this repository:

```bash
   git clone https://github.com/joshhartwig/crawler.git
```

## Usage

```bash
./crawler http://www.amazon.com 3 100
```

The parameter order is

1. The site to crawl
2. The maximum concurrent threads to use
3. The max page count (will stop if it hits that limit)

## Note

Running this on a large site will almost certainly get you rate limited, especially if you set the parameters high.