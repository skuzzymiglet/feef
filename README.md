# feef

feef is a tool for manipulating RSS/Atom/JSON feeds. It 

feef can simplify several of your other RSS-related tools down to a pipeline or a script.

feef is currently pre-alpha and constantly evolving (and breaking). 

# design

+ No unread/read tracking. Nobody can keep up with everything these days, especially the big Reddit feeds. Of course, you can script that yourself
+ No UI. Just a CLI.

# why?

Previously, I used newsboat. 

feef can be easily integrated and customized, and jobs like downloading podcast episodes or getting torrents.

# installation

`go get -u -v git.sr.ht/~skuzzymiglet/feef`

# usage

Put a file with the URLs you want to read, one per line, in `.config/feef/urls` or wherever you put your configs. This will automatically be used.
Alternatively, specify a URLs file with the `-U` flag.

+ `feef` - prints GUID of every item in every feed in your URLs file
+ `feef -u '*' -i '*{foo,bar}*'` - prints GUID of every item whose title/url matches the glob
+ `feef -f '{{.Feed.Title}}: {{.Title}}' -u '*' -i '*'` - prints every item according to a template
+ `feef 'https://100r.co/links/rss.xml' '*'` - prints GUID of every item from a feed. **it doesn't have to be in your URLs file, because it's exact**

# roadmap

(after the first quasi-stable release)

+ item-level caching (for speed)
+ concise templating - Go templates are flexible but verbose
+ configurable user agent (for example, reddit feed GET requets need a custom user-agent)
+ flexible filtering with [expr](https://github.com/antonmedv/expr)
+ server-client model for better caching

# integration

```sh
# A basic RSS reader with fzf:
# sorted by date, so everything needs to be accumulated which is slower
feef -s -f '{{.Feed.Title}}: {{.Title}} @ {{.Link}}' | fzf --bind "enter:execute(echo {} | cut -d'@' -f2 | xargs $BROWSER {})"
# random
feef -f "{{.Link}}" | shuf -n1 | xargs qutebrowser
# Notifications with notify-send:
feef -n new -c "notify-send '{{.Feed.Title}}' '{{.Title}}'"
# Download every Go Time episode (verbose atm)
feef -f "{{(index .Enclosures 0).URL}}" 'https://changelog.com/gotime/feed' | xargs wget -nc
# most recent 10
feef -s -f "{{(index .Enclosures 0).URL}}" 'https://changelog.com/gotime/feed' | head -n10 | xargs wget -nc
# random
feef -s -f "{{(index .Enclosures 0).URL}}" 'https://changelog.com/gotime/feed' | shuf -n1 | xargs wget -nc
# select some with fzf
feef -s -f "{{(index .Enclosures 0).URL}}" 'https://changelog.com/gotime/feed' | fzf -m | xargs wget -nc
```
