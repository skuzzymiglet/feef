# format

`t`: title
`d`: description
`l`: link
`p`: time published
`p(02/01 2006)`: time published, formatted
`c`: content
`q`: query syntax for feed item

lower: item
upper: feed

# query

`*`: all

`$`: random
`+`: newest
`-`: oldest
postfix above with number for n random/new/old
postfix with `*` for ordering

`feef arp242~*`
`feef arp242~+`
`feef *~+2`
`feef "dave cheney~-*"`

1 arg:  query (formatted to default)
2 args: format query

# implementation

```go
func Find(query string, v interface{}) error {
    // if query has 2 parts, run FindItems and fill v []LinkedFeedItem
    // if query has 1 part, run FindFeeds and fill v []gofeed.Feed
    // else error
}
```

# fzf

```sh
    feef "%q: %T %t" "*~*" | fzf --preview "echo {} | cut -d':' -f1 | xargs feef '%d\n%c'"
```

# notifications

```sh
feef -n -u urls "arp242"
for i in $(feef -n -u "%F %t" "lobsters");do
    notify-send $i
done
```

# roadmap

+ 
