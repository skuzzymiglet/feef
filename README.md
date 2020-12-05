# feef

feef prints and queries rss/atom/json feeds. It's very composable with other tools (through stdout)

feef is currently pre-alpha and constantly evolving (and breaking)

# installation

`go get -u -v git.sr.ht/~skuzzymiglet/feef`

Put file with the URLs you want to read, one per line, in `.config/feef/urls` or wherever you put your configs. Alternatively, specify a URLs file with the `-u` flag.

# usage

+ `feef` - prints GUID of every item in every feed in your URLs file
+ `feef '*' '*git*'` - prints every item whose title/url matches the glob

# v2 features

(after the first quasi-stable release)

+ item-level caching
+ concise templating
+ reddit user-agent
+ flexible filtering with [expr](https://github.com/antonmedv/expr)
+ feed fetching daemon

# integration

```sh
# A basic RSS reader with fzf:
# sorted by date, so everything needs to be accumulated which is slower
feef -s -f '{{.Feed.Title}}: {{.Title}} @ {{.Link}}' | fzf --bind "enter:execute(echo {} | cut -d'@' -f2 | xargs $BROWSER {})"
# random
feef -f "{{.Link}}" | shuf -n1 | xargs qutebrowser
# Notifications with notify-send:
feef -n new -c "notify-send '{{.Feed.Title}}' '{{.Title}}'"
```
