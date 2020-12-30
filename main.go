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
	"text/template"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
)

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

	// Names and stuff are a bit iffy here
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

		urlGlob  string
		itemGlob string
	)
	flag.BoolVarP(&help, "help", "h", false, "print help and exit")
	flag.StringVarP(&logLevel, "loglevel", "l", "info", "log level")
	flag.StringVarP(&urlsFile, "url-file", "U", defaultUrlsFile, "file with newline delimited URLs")
	flag.StringVarP(&templateString, "template", "f", defaultTemplate, "output template for each feed item")
	flag.StringVarP(&cmd, "exec", "c", "", "execute command template for each item")
	flag.IntVarP(&max, "max", "m", 0, "maximum items to output, 0 for no limit")
	flag.DurationVarP(&timeout, "timeout", "t", time.Second*5, "feed-fetching timeout")
	flag.IntVarP(&threads, "download-threads", "p", runtime.GOMAXPROCS(0), "maximum number of concurrent downloads") // NOTE: I'm not sure GOMAXPROCS is a reasonable default for this. Maybe we should set it to 1 for safety but that's slow
	flag.BoolVarP(&sort, "sort", "s", false, "sort feed items chronologically")
	flag.StringVarP(&notifyMode, "notify-mode", "n", "none", "notification mode (none, new or all)")
	flag.DurationVarP(&notifyPoll, "notify-poll-time", "r", 2*time.Minute, "time between feed refreshes in notification mode")

	flag.StringVarP(&urlGlob, "url-glob", "u", "*", "URL glob")
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
	if u, err := url.ParseRequestURI(urlGlob); err == nil && u.Scheme != "" { // URL, exact
		urls = []string{flag.Arg(0)}
	} else {
		var feedURL glob.Glob
		if flag.Arg(0) != "" {
			feedURL, err = glob.Compile(urlGlob)
			if err != nil {
				log.Fatalf("error compiling feed glob: %s", err)
			}
		} else {
			feedURL = glob.MustCompile("*")
		}
		if urlsFile != "" { // TODO: treatment of empty URLs parameter is a tad confusing
			file, err := os.Open(urlsFile)
			if err != nil {
				log.Fatal(err)
			}
			for _, v := range parseURLs(file) {
				if feedURL.Match(v) {
					urls = append(urls, v)
				}
			}
			file.Close()
		}
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

	var buf, cmdBuf bytes.Buffer // Buffers, so we don't print partially executed, errored templates
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
			fmt.Println() // TODO: let template choose newline. but kinda eh
		}
	}
}
