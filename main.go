// A first attempt at a simple web crawler in Go
// Future improvements:
// 1. Proper error handling
// 2. Logging
// 3. Storing of links
// 4. Customization on what to crawl for

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Start the webcrawer
func main() {
	// Starting urls
	urls := []string{"http://www.vox.com", "http://www.npr.org"}
	worklist := make(chan string)
	go buildWorklist(urls, worklist)
	go crawlWorklist(worklist)

	// To fix with an actual termination condition
	time.Sleep(1 * time.Minute)

}

// buildWorklist takes a string slice and sends each to worklist channel
func buildWorklist(links []string, worklist chan string) {
	for _, url := range links {
		worklist <- url
	}
}

// Channel to restrict number of visited URLs
var tokens = make(chan struct{}, 20)

// crawlWorklist takes the urls from worklist to send the resulting link slice
// back into the worklist via buildWorklist
func crawlWorklist(worklist chan string) {
	for url := range worklist {
		tokens <- struct{}{} // acquire a token for each url to be processed
		go func(url string) {
			links := ExtractLinks(url)
			<-tokens // release the token once the url is processed
			buildWorklist(links, worklist)
		}(url)
	}
}

// ExtractLinks requests the url and returns all http links embedded
func ExtractLinks(url string) (links []string) {
	// Request the URL and parse it to return a document tree
	r, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	doc, err := html.Parse(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	// visitNode recursively visits each node in the document
	var visitNode func(*html.Node)
	visitNode = func(n *html.Node) {
		// Select the nodes that are links
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					if strings.HasPrefix(attr.Val, "http") {
						// Remove white space chars
						link := strings.TrimSpace(attr.Val)
						fmt.Println(link)
						links = append(links, link)
					}
					continue
				}
			}
		}

		// For loop to visit every node in the document breadth first
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visitNode(c)
		}

	}
	visitNode(doc)
	return links
}
