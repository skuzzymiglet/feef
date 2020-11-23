# feef

feef prints and queries rss/atom/json feeds. It's very composable with other tools (through stdout)

# installation

`go get -u -v git.sr.ht/~skuzzymiglet/feef`

Put URLs you want to read in `.config/feef/urls` or wherever you put your configs. Alternatively, specify one with the `-u` flag.

`feef` - prints GUID of every item of the feeds in urls
`feef "*" "*git*"` - prints every item whose title/url matches the glob

# integration

```sh
# A basic RSS reader with fzf:
# sorted by date, so everything needs to be accumulated which is slower
feef -s -f '{{.Feed.Title}}: {{.Title}} @ {{.Link}}' | fzf --bind "enter:execute(echo {} | cut -d'@' -f2 | xargs $BROWSER {})"
# random
feef -f "{{.Link}}" | shuf -n1 | xargs qutebrowser
# Notifications with notify-send:
feef -n -c "notify-send '{{.Feed.Title}}' '{{.Title}}'"
# Download every Go Time episode
feef -f "{{.PodcastDownload}}" "https://changelog.com/gotime/feed" "*" | xargs wget -nc
```

# todo

+ caching
+ more detailed queries
+ internal architecture cleanup
