package main

import (
	"context"
	"time"
)

type NotifyParam struct {
	urls        []string
	poll        time.Duration
	maxDownload int
}

func NotifyNew(ctx context.Context, n NotifyParam, out chan LinkedFeedItem, errChan chan error) {
	sema := make(chan struct{}, n.maxDownload)
	for _, u := range n.urls {
		// TODO: stagger refreshes to reduce semaphore contention,maybe
		initial := true
		var last LinkedFeed
		go func(u string) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					sema <- struct{}{}
					// logrus.Debugln("refreshing", u)
					lf, err := Get(u)
					if err != nil {
						errChan <- err
						continue
					}
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
