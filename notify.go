package main

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type NotifyMode int

const (
	newItems NotifyMode = iota
	allItems
)

type NotifyParam struct {
	urls        []string
	mode        NotifyMode
	poll        time.Duration
	maxDownload int
}

func NotifyNew(ctx context.Context, n NotifyParam, out chan LinkedFeedItem, errChan chan error) {
	sema := make(chan struct{}, n.maxDownload)
	for _, u := range n.urls {
		// When only showing new items, fetch the initial feed
		// Othwerwise start with nothing
		initial := n.mode == newItems

		var last LinkedFeed
		// var lastTime time.Time // 0, initially download everything
		go func(u string) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					sema <- struct{}{}
					log.Debugln("refreshing", u)

					// Issue a HEAD
					// Check etag/last-modified
					// Pass to parser
					// LinkFeed
					lf, err := Get(u) // TODO: refactor Get so we can issue a HEAD request for caching, then make a request with a context on a http client. Probably make it just parse a reader lol
					if err != nil {
						errChan <- err
						continue
					}

					// 					if v, ok := lf.Headers["last-modified"]; ok {
					// 						if t, err := time.Parse(time.RFC1123, v); err == nil {
					// 							if t.Before(lastTime) { // No changes
					// 								log.Debugln("")
					// 								continue
					// 							}
					// 						}
					// 					}
					if initial { // Don't compare
						initial = false
					} else {
						for _, i := range lf.Items { // For each new...
							matched := false
							for _, j := range last.Items { // Compare each old
								if i.GUID == j.GUID {
									matched = true
								}
							}
							if !matched {
								log.Debugf("found new item %s", i.GUID)
								out <- i
							}
						}
					}
					last = lf
					<-sema
					time.Sleep(n.poll)
				}
			}
		}(u)
	}
}
