package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/mmcdole/gofeed"
)

func main() {
	// Command line syntax:
	// feef query format
	urls := flag.String("u", "urls", "file with newline delimited URLs")
	query := flag.String("q", "", "query (feed~item)")
	format := flag.String("f", "%t %u", "format for printing")
	flag.Parse()
	switch os.Args[1] {
	case "list":
		urlsFile, err := os.Open(*urls)
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(urlsFile)
		var wg sync.WaitGroup
		for scanner.Scan() {
			wg.Add(1)
			go func(u string, wg *sync.WaitGroup) {
				parser := gofeed.NewParser() // fucking race conditions
				defer wg.Done()
				resp, err := http.Get(u)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error downloading %s: %s\n", u, err)
					return
				}
				defer resp.Body.Close()
				feed, err := parser.Parse(resp.Body)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error parsing %s: %s\n", u, err)
					return
				}
				lf := LinkFeed(feed)
				fmt.Printf(*format, lf)
				fmt.Println()
			}(scanner.Text(), &wg)
		}
		wg.Wait()
	case "feeds":
	case "show":
		item, err := FindItem(*query)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf(*format, item)
		fmt.Println()
	}
}
