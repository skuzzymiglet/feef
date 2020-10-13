# tabs

# new 

- panes contain 1 item.
- Most recent unread item that's newer than 3 days (configurable)

# unread

- like new but all unread items

# search 

- searchbox (for all feeds)
+ input box on top (100% width)
+ left box: metadata
+ right box: content

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
+ tab - next item
+ enter - next feed

**users are not expected to read items in the application (there will be no super cool reading program like newsboat, an external program should be used). they will be presented**

# storage

## read/unread

### plaintext file

+ easy to awk
+ _hash_:_date_:_feed url_:_title_:_url_:_podcast download url_
+ delimiter should be very obscure, but consistent (for awking)

### sqlite

+ relational
+ difficult to write, ehh for scripts

### kv db

+ simple data model
+ `feed url:guid/ item url`: `bool` (0 or 1, nice and simple)

## cache

+ not necessary really, except for offline reading
+ load at start, write at end, so simple dump may work

# subcommands

ideally, feef will integrate with fzf/dmenu and other programs, by providing useful output

+ none: start ui
+ list: list 
