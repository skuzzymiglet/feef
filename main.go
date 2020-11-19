package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"
)

func printHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: \n\n%s [format] [query]\n\n", os.Args[0], os.Args[0]) // TODO: make this tidier
	flag.PrintDefaults()
}
func main() {
	defaultTemplate := "{{.Title}}: {{.Link}} ({{.Feed.Title}})"
	// Command line syntax:
	// feef query format
	urlsFile := flag.String("u", "urls", "file with newline delimited URLs")
	templateString := flag.String("f", defaultTemplate, "output template for each feed item")
	help := flag.Bool("h", false, "print help and exit")
	flag.Parse()

	if *help {
		printHelp()
	}

	// Parse template
	tmpl, err := template.New("output").Parse(*templateString)
	if err != nil {
		log.Fatal(err)
	}

	// Get list of URLs
	urls := make([]string, 0)
	file, err := os.Open(*urlsFile)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	// TODO: QUERIES
	// + use gobwas/glob
	// + have 2 args rather than `x~y` cos urls can contain ~
	switch flag.NArg() {
	case 1: // Query
		var v []LinkedFeedItem
		err := Find(flag.Arg(0), &v, urls)
		if err != nil {
			log.Fatal(err)
		}
		for _, val := range v {
			err := tmpl.Execute(os.Stdout, val)
			if err != nil {
				log.Fatal(err)
			}
			os.Stdout.Write([]byte("\n")) // Do we need to check this?
		}
	default:
		printHelp()
		os.Exit(2)
	}
}
