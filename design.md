# tabs

## new

+ central pane
+ others tiled behind
+ tab to switch

## old

tree?

feed>item

right pane: view of item/feed

## queue

+ one pane per feed (with title in top)
+ panes contain: 
    + shortest date possible
    + title
    + author (if there are multiple different authors)

## jobs

list of commands:

+ errored: red (show status code)
+ exited: gray (exit code 0)
+ running: green

j/k to move up/down

right pane: output (if there's too much, the bottom is shown)

# keys

## item/feed local

(capitalized equilavents use the podcast download URL)

+ `o` - spawn `open` script for URL
+ `b` - open URL in browser
+ `y` - yank URL

## global

+ `q` - exit
+ `r` - reload all feeds
+ `C-r` - redraw screen

users are not expected to read items in the application (there will be no super cool reading program like newsboat, an external program should be used). they will be presented 

# how to store read items

plaintext file (that is easy to awk)

_hash_:_date_:_feed url_:_title_:_url_:_podcast download url_

(delimiter should be very obscure, but consistent (for awking))

sqlite (yuck, but maybe)

# subcommands

ideally, feef will integrate with fzf/dmenu and other programs, by providing useful output

+ none: start ui
+ list: list 
