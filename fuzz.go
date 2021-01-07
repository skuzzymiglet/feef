package main

import (
	"time"

	"github.com/brianvoe/gofakeit/v5"
	"github.com/gorilla/feeds"
)

func makeFeed(start time.Time) feeds.Feed {
	itemsN := int(time.Now().Sub(start).Seconds())
	items := make([]*feeds.Item, itemsN)
	for i := range items {
		items[i] = &feeds.Item{
			Title:       gofakeit.BS(),
			Link:        &feeds.Link{Href: gofakeit.URL()},
			Description: gofakeit.LoremIpsumSentence(10),
			Author:      &feeds.Author{Name: gofakeit.Name(), Email: gofakeit.Email()},
			Created:     time.Now().Add(time.Duration(-i) * time.Second),
		}
	}
	return feeds.Feed{
		Title:       gofakeit.BS(),
		Link:        &feeds.Link{Href: gofakeit.URL()},
		Description: gofakeit.LoremIpsumSentence(10),
		Author:      &feeds.Author{Name: gofakeit.Name(), Email: gofakeit.Email()},
		Created:     time.Now(),
		Items:       items,
	}
}
