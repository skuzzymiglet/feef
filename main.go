package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
)

func printHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: \n\n%s [feed URL glob] [item glob]\n\n", os.Args[0], os.Args[0]) // TODO: make this tidier
	flag.PrintDefaults()
}

var defaultFuncMap = map[string]interface{}{
	"date": func(t time.Time) string {
		return t.Format("January 2, 2006")
	},
	"format": func(fmt string, t time.Time) string {
		return t.Format(fmt)
	},
}

var tmpl = template.New("output").
	Funcs(defaultFuncMap)

var cmdTmpl = template.New("cmd").
	Funcs(defaultFuncMap)

var defaultTemplate = "{{.GUID}}"

func main() {
	var defaultUrlsFile string
	cdir, err := os.UserConfigDir()
	if err == nil {
		defaultUrlsFile = filepath.Join(cdir, "feef", "urls")
	}

	help := flag.Bool("h", false, "print help and exit")

	logLevel := flag.String("l", "info", "log level")

	urlsFile := flag.String("u", defaultUrlsFile, "file with newline delimited URLs")
	templateString := flag.String("f", defaultTemplate, "output template for each feed item")
	cmd := flag.String("c", "", "execute command template for each item")

	max := flag.Int("m", 0, "maximum items to output, 0 for no limit")
	timeout := flag.Duration("t", time.Second*5, "feed-fetching timeout")
	// I'm not sure GOMAXPROCS is a reasonable default for this. Maybe we should set it to 1 for safety but that's slow
	threads := flag.Int("p", runtime.GOMAXPROCS(0), "maximum number of concurrent downloads")
	sort := flag.Bool("s", false, "sort feed items chronologically")

	notifyMode := flag.String("n", "none", "notification mode (none, new or all)")
	notifPoll := flag.Duration("r", 2*time.Minute, "time between feed refreshes in notification mode")

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)

	// Parse output template
	_, err = tmpl.Parse(*templateString)
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}
	if *cmd != "" {
		_, err = cmdTmpl.Parse(*cmd)
		if err != nil {
			log.Fatal("error parsing command template: %s", err)
		}
	}

	// Get list of URLs
	/*
	   TODO: perhaps we should scan every file under .config/feef/urls (directory)
	   people may want to separate youtube URLs from, say reddit
	*/
	urls := make([]string, 0)
	if u, err := url.ParseRequestURI(flag.Arg(0)); err == nil && u.Scheme != "" { // URL, exact
		urls = []string{flag.Arg(0)}
	} else {
		var feedURL glob.Glob
		if flag.Arg(0) != "" {
			feedURL, err = glob.Compile(flag.Arg(0))
			if err != nil {
				log.Fatalf("error compiling feed glob: %s", err)
			}
		} else {
			feedURL = glob.MustCompile("*")
		}
		if *urlsFile != "" { // TODO: treatment of empty URLs parameter is a tad confusing
			file, err := os.Open(*urlsFile)
			if err != nil {
				log.Fatal(err)
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				if strings.TrimSpace(scanner.Text()) != "" {
					if !strings.HasPrefix(scanner.Text(), "#") {
						if feedURL.Match(scanner.Text()) {
							urls = append(urls, scanner.Text())
						}
					}
				}
			}
			file.Close()
		}
	}

	items := make(chan LinkedFeedItem)
	results := make(chan LinkedFeedItem)
	errChan := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())

	p := GetParam{
		client:     &http.Client{Timeout: *timeout},
		maxThreads: *threads,
		urls:       urls,
	}
	np := NotifyParam{
		GetParam: p,
		poll:     *notifPoll,
	}

	switch *notifyMode {
	case "none":
		go func() {
			Get(ctx, p, items, errChan)
			close(items)
		}()
	case "new":
		np.mode = newItems
		go func() {
			Notify(ctx, np, items, errChan)
			close(items)
		}()
	case "all":
		np.mode = allItems
		go func() {
			Notify(ctx, np, items, errChan)
			close(items)
		}()
	default:
		log.Fatalf("Invalid notify mode %s", *notifyMode)
	}

	// filtering
	fp := FilterParam{max: *max, sort: *sort, item: glob.MustCompile("*")}
	if flag.Arg(1) != "" {
		item, err := glob.Compile(flag.Arg(1))
		if err != nil {
			log.Fatal("error compiling item glob: ", err)
		}
		fp.item = item
	}
	if *notifyMode != "none" {
		log.Warn("Sorting in notify mode blocks forever, disabling sorting")
		fp.sort = false // Sorting would block forever
	}
	go func() {
		Filter(fp, items, results, errChan)
		close(results)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, die...)

	var buf, cmdBuf bytes.Buffer // Buffers, so we don't print partially executed, errored templates
	for {
		select {
		case s := <-c:
			log.Fatalf("Got signal: %s", s)
			cancel()
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
						log.Errorf("error running command %s (%s)", xe, string(xe.Stderr)) // BUG: doesn't show stderr
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
			fmt.Println()
		}
	}
}
