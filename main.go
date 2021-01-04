package main

/*
NOTE: main.go is really complex
Target: reduce it too 100ish lines by v1
*/

import (
	"bytes"
	"context"
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
	"time"

	flag "github.com/spf13/pflag"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
)

func main() {
	var defaultUrlsFile string
	cdir, err := os.UserConfigDir()
	if err == nil {
		defaultUrlsFile = filepath.Join(cdir, "feef", "urls")
	}

	// Names and stuff are a bit inconsistent here
	var (
		help           bool
		logLevel       string
		urlsFile       string
		templateString string
		cmd            string
		max            int
		timeout        time.Duration
		threads        int
		sort           bool
		notifyMode     string
		notifyPoll     time.Duration

		urlGlobs []string
		itemGlob string
	)
	flag.BoolVarP(&help, "help", "h", false, "print help and exit")
	flag.StringVarP(&logLevel, "loglevel", "l", "info", "log level")
	flag.StringVarP(&urlsFile, "url-file", "U", defaultUrlsFile, "file with newline delimited URLs")
	flag.StringVarP(&templateString, "template", "f", defaultTemplate, "output template for each feed item")
	flag.StringVarP(&cmd, "exec", "c", "", "execute command template for each item")
	// TODO: template to run with a slice of all items (webrings)
	flag.IntVarP(&max, "max", "m", 0, "maximum items to output, 0 for no limit")
	flag.DurationVarP(&timeout, "timeout", "t", time.Second*5, "feed-fetching timeout")
	flag.IntVarP(&threads, "download-threads", "p", runtime.GOMAXPROCS(0), "maximum number of concurrent downloads") // NOTE: I'm not sure GOMAXPROCS is a reasonable default for this. Maybe we should set it to 1 for safety but that's slow
	flag.BoolVarP(&sort, "sort", "s", false, "sort feed items chronologically")
	flag.StringVarP(&notifyMode, "notify-mode", "n", "none", "notification mode (none, new or all)")
	flag.DurationVarP(&notifyPoll, "notify-poll-time", "r", 2*time.Minute, "time between feed refreshes in notification mode")

	flag.StringSliceVarP(&urlGlobs, "url-glob", "u", []string{"*"}, "URLs or URL globs matched against URLs file")
	// This glob stuff is silly
	// https://* evaluates to a URL
	// TODO: We should do prefixes:
	// ~ fuzzy
	// ? glob
	// / regexp
	// <nothing> for exact
	// ambiguity is the devil's work
	flag.StringVarP(&itemGlob, "item-glob", "i", "*", "item glob")

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)

	// Parse output template
	_, err = tmpl.Parse(templateString)
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}
	if cmd != "" {
		_, err = cmdTmpl.Parse(cmd)
		if err != nil {
			log.Fatalf("error parsing command template: %s", err)
		}
	}

	// Get list of URLs
	/*
			   TODO: perhaps we should scan every file under .config/feef/urls (directory)
			   people may want to separate youtube URLs from, say reddit
		       TODO: fetch newly added feed URLs (fsnotify)
	*/
	urls := make([]string, 0)
	var urlsFileURLs []string
	if urlsFile != "" { // TODO: treatment of empty URLs parameter is a tad confusing

		file, err := os.Open(urlsFile)
		if err != nil {
			log.Fatal(err)
		}
		urlsFileURLs = parseURLs(file)
		file.Close()
	}
	for _, urlGlob := range urlGlobs {
		if u, err := url.ParseRequestURI(urlGlob); err == nil && u.Scheme != "" { // URL provided is exact, no matching needed
			urls = append(urls, urlGlob)
		} else {
			if len(urlsFileURLs) == 0 {
				log.Fatalf("URL glob '%s' provided but no URLs to match against", urlGlob)
			}
			var feedURL glob.Glob
			feedURL, err = glob.Compile(urlGlob)
			if err != nil {
				log.Fatalf("error compiling feed glob: %s", err)
			}
			matchedOne := false
			for _, v := range urlsFileURLs {
				if feedURL.Match(v) {
					matchedOne = true
					urls = append(urls, v)
				}
			}
			if !matchedOne {
				log.Warnf("URL glob %s matched no URLs in %s", urlGlob, urlsFile)
			}
		}
	}
	if len(urls) == 0 {
		log.Fatalf("No URLs matched")
	}

	items := make(chan LinkedFeedItem, 1)
	results := make(chan LinkedFeedItem, 1)
	errChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())

	p := GetParam{
		client:     &http.Client{Timeout: timeout}, // this is deprecated, TODO use context, I think
		maxThreads: threads,
		urls:       urls,
	}
	np := NotifyParam{
		GetParam: p,
		poll:     notifyPoll,
	}

	// TODO: notifymode struct, which satisfies flag.Value
	switch notifyMode {
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
		log.Fatalf("Invalid notify mode %s", notifyMode)
	}

	// filtering
	// TODO: flexible filtering with expr
	fp := FilterParam{max: max, sort: sort, item: glob.MustCompile("*")}

	item, err := glob.Compile(itemGlob)
	if err != nil {
		log.Fatal("error compiling item glob: ", err)
	}
	fp.item = item

	if notifyMode != "none" {
		log.Warn("Sorting in notify mode blocks forever, disabling sorting")
		fp.sort = false // Sorting would block forever
	}
	go func() {
		Filter(fp, items, results, errChan)
		close(results)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, die...) // BUG: SIGPIPEs are not handled in notify new mode (strange)

	var (
		// stdout is bytes, only io.Copy is used
		buf bytes.Buffer // Buffers, so we don't print partially executed, errored templates
		// command is string, only every read as string
		cmdBuf strings.Builder
	)
	for {
		select {
		case s := <-c:
			log.Errorf("Got signal: %s, cancelling", s)
			cancel()
			os.Exit(1)
		case err := <-errChan:
			log.Error(err)
		case val, more := <-results:
			if !more {
				return
			}
			if cmd != "" {
				cmdTmpl.Execute(&cmdBuf, val)
				cmd := exec.Command("sh", "-c", cmdBuf.String())
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					switch exitError := err.(type) {
					case *exec.ExitError:
						log.Errorf("error running command %s: %s", cmdBuf.String(), exitError)
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
			fmt.Println() // TODO: let template choose newline. but kinda eh
		}
	}
}
