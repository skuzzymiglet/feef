package main

import (
	"bufio"
	"io"
	"strings"
)

func parseURLs(r io.Reader) (urls []string) {
	// Q: should we validate the URLs here?
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			if !strings.HasPrefix(scanner.Text(), "#") { // TODO: comments at ends of lines
				urls = append(urls, scanner.Text())
			}
		}
	}
	return
}
