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
	// defaultTemplate := "{{.Title}}: {{.Link}} ({{.Feed.Title}})"
	defaultTemplate := "{{.Feed.Link}} {{.Link}}"
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
	if flag.NArg() != 2 {
		printHelp()
		os.Exit(2)
	}
	v, err := FindItems(flag.Arg(0), flag.Arg(1), urls)
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
}
