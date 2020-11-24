package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/gobwas/glob"
	"github.com/sirupsen/logrus"
)

func printHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: \n\n%s [format] [query]\n\n", os.Args[0], os.Args[0]) // TODO: make this tidier
	flag.PrintDefaults()
}
func main() {
	logrus.SetLevel(logrus.DebugLevel) // TODO: fully switch to logrus

	var defaultUrlsFile string
	cdir, err := os.UserConfigDir()
	if err == nil {
		defaultUrlsFile = filepath.Join(cdir, "feef", "urls")
	}

	defaultTemplate := "{{.GUID}}"

	help := flag.Bool("h", false, "print help and exit")

	urlsFile := flag.String("u", defaultUrlsFile, "file with newline delimited URLs")
	templateString := flag.String("f", defaultTemplate, "output template for each feed item")
	cmd := flag.String("c", "", "execute command template for each item")

	max := flag.Int("m", 0, "maximum items to output, 0 for no limit") // BUG: shows nothing. needs diagnosing
	threads := flag.Int("p", runtime.GOMAXPROCS(0), "maximum number of concurrent downloads")
	sort := flag.Bool("s", false, "sort by when published")

	notify := flag.Bool("n", false, "print new items as they're published")
	notifPoll := flag.Duration("r", time.Second*10, "time between feed refreshes in notification mode")

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	// Parse output template
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
	var cmdTmpl *template.Template
	if *cmd != "" {
		cmdTmpl, err = template.New("cmd").
			Funcs(map[string]interface{}{
				"date": func(t time.Time) string {
					return t.Format("January 2, 2006")
				},
				"format": func(fmt string, t time.Time) string {
					return t.Format(fmt)
				},
			}).
			Parse(*cmd)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Get list of URLs
	urls := make([]string, 0)
	if *urlsFile != "" {
		file, err := os.Open(*urlsFile)
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if !strings.HasPrefix(scanner.Text(), "#") {
				urls = append(urls, scanner.Text())
			}
		}
		file.Close()
	}
	results := make(chan LinkedFeedItem, 100)
	errChan := make(chan error)

	p := Param{
		max:     *max,
		urls:    urls,
		sort:    *sort,
		item:    glob.MustCompile("*"),
		feedURL: glob.MustCompile("*"),
	}
	switch flag.NArg() {
	case 2:
		feedURL, err := glob.Compile(flag.Arg(0))
		if err != nil {
			log.Fatal("error compiling feed URL glob: ", err)
		}
		p.feedURL = feedURL

		item, err := glob.Compile(flag.Arg(1))
		if err != nil {
			log.Fatal("error compiling item glob: ", err)
		}
		p.item = item

	case 1:
		parts := strings.Split(flag.Arg(0), delim)
		if len(parts) != 2 {
			log.Fatalf("Not enought parts in query %s", flag.Arg(0))
		}
		feedURL, err := glob.Compile(parts[0])
		if err != nil {
			log.Fatal("error compiling feed URL glob: ", err)
		}
		p.feedURL = feedURL

		item, err := glob.Compile(parts[1])
		if err != nil {
			log.Fatal("error compiling item glob: ", err)
		}
		p.item = item
	}
	go func() {
		items := make(chan LinkedFeedItem, 0)
		if *notify {
			ctx := context.Background()
			go NotifyNew(ctx, NotifyParam{urls: urls, poll: *notifPoll, maxDownload: *threads}, items, errChan)
		} else {
			go func() {
				GetAll(p.urls, *threads, items, errChan)
				close(items)
			}()
		}
		Filter(p, items, results, errChan)
		close(results)
	}()
	var buf, cmdBuf bytes.Buffer
	for {
		select {
		case err := <-errChan:
			log.Println(err)
		case val, more := <-results:
			if !more {
				return
			}
			if *cmd != "" {
				cmdTmpl.Execute(&cmdBuf, val)
				cmd := exec.Command("sh", "-c", string(cmdBuf.Bytes()))
				cmd.Stdout = os.Stdout
				err := cmd.Run()
				if err != nil {
					switch xe := err.(type) {
					case *exec.ExitError:
						log.Printf("error running command %s (%s)", xe, string(xe.Stderr))
					default:
						log.Fatal(err)
					}
				}
				cmdBuf.Reset()
			}
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
	}
}
