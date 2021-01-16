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
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/peterbourgon/diskv"
	"github.com/pkg/profile"
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

	var client *http.Client
	var itemCache *diskv.Diskv
	cacheDir, err := os.UserCacheDir()
	if err == nil {
		err := os.Mkdir(filepath.Join(cacheDir, "feef"), 0640)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		cache := diskcache.New(filepath.Join(cacheDir, "feef", "http"))
		itemCache = diskv.New(diskv.Options{BasePath: filepath.Join(cacheDir, "feef", "items")})
		client = &http.Client{Transport: httpcache.NewTransport(cache)}
	}

	// Names and stuff are a bit inconsistent here
	var (
		help                bool
		logLevel            string
		urlsFile            string
		templateString      string
		cmd                 string
		max                 int
		timeout             time.Duration
		threads             int
		sort                bool
		notifyMode          string
		notifyPoll          time.Duration
		exitOnFailedCommand bool
		exitOnFailedFetch   bool

		urlSpecs []string
		itemGlob string

		memProfile bool
		cpuProfile bool
	)
	flag.BoolVarP(&help, "help", "h", false, "print help and exit")
	flag.StringVarP(&logLevel, "loglevel", "l", "info", "log level")
	flag.StringVarP(&urlsFile, "url-file", "U", defaultUrlsFile, "file with newline delimited URLs")
	flag.StringVarP(&templateString, "template", "f", defaultTemplate, "output template for each feed item")
	flag.StringVarP(&cmd, "exec", "c", "", "execute command template for each item")
	// TODO: template to run with a slice of all items (webrings)
	flag.IntVarP(&max, "max", "m", 0, "maximum items to output, 0 for no limit")
	flag.DurationVarP(&timeout, "timeout", "t", time.Second*5, "feed-fetching timeout")
	flag.BoolVarP(&exitOnFailedCommand, "exit-on-failed-command", "e", false, "exit if a command (-c) fails")
	flag.BoolVarP(&exitOnFailedFetch, "exit-on-failed-fetch", "E", false, "exit if fetching a feed fails (4xx or 5xx response code)")
	flag.IntVarP(&threads, "download-threads", "p", runtime.GOMAXPROCS(0), "maximum number of concurrent downloads") // NOTE: I'm not sure GOMAXPROCS is a reasonable default for this. Maybe we should set it to 1 for safety but that's slow
	flag.BoolVarP(&sort, "sort", "s", false, "sort feed items chronologically")
	flag.StringVarP(&notifyMode, "notify-mode", "n", "none", "notification mode (none, new or all)")
	flag.DurationVarP(&notifyPoll, "notify-poll-time", "r", 2*time.Minute, "time between feed refreshes in notification mode")

	// NOTE: pflag doesn't let you re-specify flags, which is more foolproof than splitting by ','. Maybe getopt?
	flag.StringSliceVarP(&urlSpecs, "url-spec", "u", []string{"~"}, "List of URLs or URL patterns to match against the URLs file (prefixes: / for regexp, ~ for fuzzy match, ? for glob)") // poor documentation
	flag.StringVarP(&itemGlob, "item-glob", "i", "*", "item glob")

	flag.BoolVar(&memProfile, "memory-profile", false, "record memory profile")
	flag.BoolVar(&cpuProfile, "cpu-profile", false, "record CPU profile")

	flag.Parse()

	client.Timeout = timeout

	switch {
	case memProfile:
		defer profile.Start(profile.MemProfile).Stop()
	case cpuProfile:
		defer profile.Start().Stop()
	}

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
	var (
		urls         []string
		urlsFileURLs []string
	)
	if urlsFile != "" { // TODO: treatment of empty URLs parameter is a tad confusing
		file, err := os.Open(urlsFile)
		if err != nil {
			log.Fatal(err)
		}
		urlsFileURLs = parseURLs(file)
		file.Close()
	}
	for _, urlSpec := range urlSpecs {
		u, err := matchURLSpec(urlSpec, urlsFileURLs)
		if err != nil {
			log.Fatalf("Error parsing URL spec: %s", err)
		}
		if len(u) == 0 {
			log.Warnf("URL spec %s matched no URLs in %s", urlSpec, urlsFile)
		}
		urls = append(urls, u...)
	}
	if len(urls) == 0 {
		log.Fatalf("No URLs matched")
	}

	items := make(chan LinkedFeedItem, 1)
	results := make(chan LinkedFeedItem, 1)
	errChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())

	p := GetParam{
		urls: urls,
		Fetcher: Fetcher{
			client:    client,
			itemCache: itemCache,
			sema:      make(chan struct{}, threads),
		},
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

	// Buffers, so we don't print partially executed, errored templates
	var (
		tmplBuf bytes.Buffer
		cmdBuf  strings.Builder
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
						if exitOnFailedCommand {
							log.Fatalf("error running command %s: %s", cmdBuf.String(), exitError)
						} else {
							log.Errorf("error running command %s: %s", cmdBuf.String(), exitError)
						}
					default:
						log.Fatal(err)
					}
				}
				cmdBuf.Reset()
			}
			err := tmpl.Execute(&tmplBuf, val)
			if err != nil {
				log.Fatalln("error executing template:", err)
				fmt.Printf("(ERROR)")
			} else {
				io.Copy(os.Stdout, &tmplBuf)
				tmplBuf.Reset()
			}
			fmt.Println() // TODO: let template choose newline. but kinda eh
		}
	}
}
