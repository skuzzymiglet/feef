package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"github.com/lithammer/fuzzysearch/fuzzy"
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

// matchURLSpec parses a single URL specification
// Prefixes
// ~: fuzzy-matched against URLs
// /: regexp
// ?: glob (not quite sure)
// otherwise, parsed as URL
func matchURLSpec(spec string, urls []string) ([]string, error) {
	if len(spec) == 0 {
		return []string{}, errors.New("Empty URL spec")
	}
	switch {
	case strings.HasPrefix(spec, "~"):
		if len(urls) == 0 {
			return []string{}, fmt.Errorf("fuzzy URL'%s' provided but no URLs to match against", spec)
		}
		return fuzzy.Find(strings.TrimPrefix(spec, "~"), urls), nil
	case strings.HasPrefix(spec, "/"):
		if len(urls) == 0 {
			return []string{}, fmt.Errorf("URL regex '%s' provided but no URLs to match against", spec)
		}
		re, err := regexp.Compile(strings.TrimPrefix(spec, "/"))
		if err != nil {
			return []string{}, err
		}
		var matches []string
		for _, u := range urls {
			if re.MatchString(u) {
				matches = append(matches, u)
			}
		}
		return matches, nil
	case strings.HasPrefix(spec, "?"):
		if len(urls) == 0 {
			return []string{}, fmt.Errorf("URL glob'%s' provided but no URLs to match against", spec)
		}
		g, err := glob.Compile(strings.TrimPrefix(spec, "?"))
		if err != nil {
			return []string{}, err
		}
		var matches []string
		for _, u := range urls {
			if g.Match(u) {
				matches = append(matches, u)
			}
		}
		return matches, nil
	}
	// exact
	_, err := url.Parse(spec)
	if err != nil {
		return []string{}, err
	}
	return []string{spec}, nil
}
