package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
)

func printHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: \n\n%s [format] [query]\n\n", os.Args[0], os.Args[0]) // TODO: make this tidier
	flag.PrintDefaults()
}

var tmpl = template.New("output").
	Funcs(map[string]interface{}{
		"date": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"format": func(fmt string, t time.Time) string {
			return t.Format(fmt)
		},
	})

var cmdTmpl = template.New("cmd").
	Funcs(map[string]interface{}{
		"date": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"format": func(fmt string, t time.Time) string {
			return t.Format(fmt)
		},
	})

func main() {
	log.SetLevel(log.InfoLevel)

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

	notifyMode := flag.String("n", "none", "notification mode (none, new or all)")
	notifPoll := flag.Duration("r", time.Second*10, "time between feed refreshes in notification mode")

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	// Parse output template
	_, err = tmpl.Parse(*templateString)
	if err != nil {
		log.Fatal(err)
	}
	if *cmd != "" {
		_, err = cmdTmpl.Parse(*cmd)
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
	}
	go func() {
		items := make(chan LinkedFeedItem, 0)
		if *notifyMode != "none" {
			ctx := context.Background()
			p := NotifyParam{urls: urls, poll: *notifPoll, maxDownload: *threads}
			switch *notifyMode {
			case "new":
				p.mode = newItems
			case "all":
				p.mode = allItems
			default:
				log.Fatalf("invalid notification mode %s", *notifyMode)
			}
			go NotifyNew(ctx, p, items, errChan)
		} else {
			go func() {
				GetAll(p.urls, *threads, items, errChan)
				close(items)
			}()
		}
		if *notifyMode != "none" {
			log.Warn("Sorting in notify mode blocks forever, disabling sorting")
			p.sort = false // Sorting would block forever
		}
		Filter(p, items, results, errChan)
		close(results)
	}()
	var buf, cmdBuf bytes.Buffer
	for {
		select {
		case err := <-errChan:
			log.Error(err)
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
						log.Errorf("error running command %s (%s)", xe, string(xe.Stderr))
					default:
						log.Fatal(err)
					}
				}
				cmdBuf.Reset()
			}
			err := tmpl.Execute(&buf, val)
			if err != nil {
				log.Errorln("error executing template:", err)
				fmt.Printf("(ERROR)")
			} else {
				io.Copy(os.Stdout, &buf)
				buf.Reset()
			}
			os.Stdout.Write([]byte("\n")) // Do we need to check this?
		}
	}
}
