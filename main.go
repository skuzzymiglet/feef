package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

func printHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: \n\n%s [format] [query]\n\n", os.Args[0], os.Args[0]) // TODO: make this tidier
	flag.PrintDefaults()
}
func main() {
	// defaultTemplate := "{{.Title}}: {{.Link}} ({{.Feed.Title}})"
	defaultTemplate := "{{.Feed.FeedLink}}" + delim + "{{.GUID}}"
	urlsFile := flag.String("u", "", "file with newline delimited URLs")
	templateString := flag.String("f", defaultTemplate, "output template for each feed item")
	help := flag.Bool("h", false, "print help and exit")
	flag.Parse()

	if *help {
		printHelp()
	}

	// Parse template
	tmpl, err := template.New("output").
		Funcs(map[string]interface{}{
			"date": func(t time.Time) string {
				return t.Format("January 2, 2006")
			},
			"format": func(fmt string, t time.Time) string {
				return t.Format(fmt)
			},
		}).
		Parse(*templateString)
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
	var v []LinkedFeedItem
	switch flag.NArg() {
	case 2:
		v, err = FindItems(flag.Arg(0), flag.Arg(1), urls)
	case 1:
		parts := strings.Split(flag.Arg(0), delim)
		v, err = FindItems(parts[0], parts[1], urls)
	}
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	for _, val := range v {
		err := tmpl.Execute(&buf, val)
		if err != nil {
			log.Println("error executing template:", err)
			fmt.Printf("(ERROR)")
		} else {
			io.Copy(os.Stdout, &buf)
			buf.Reset()
		}
		os.Stdout.Write([]byte("\n")) // Do we need to check this?
	}
	fmt.Println(len(v))
}
