package main

import (
	"text/template"

	"time"

	"github.com/gosimple/slug"
	"github.com/mattn/go-runewidth"
	"github.com/microcosm-cc/bluemonday"
)

var defaultFuncMap = map[string]interface{}{
	"datef": func(fmt string, t time.Time) string {
		return t.Format(fmt)
	},
	"slug": func(s string) string {
		return slug.Make(s)
	},
	"trunc": func(n int, s string) string {
		return runewidth.Truncate(s, n, "")
	},
	"truncPad": func(n int, s string) string {
		return runewidth.FillRight(runewidth.Truncate(s, n, ""), n)
	},
	// maybe multiple policies?
	"sanitizeHTML": func(s string) string {
		return bluemonday.StrictPolicy().Sanitize(s)
	},
}

var tmpl = template.New("output").
	Funcs(defaultFuncMap)

var cmdTmpl = template.New("cmd").
	Funcs(defaultFuncMap)

var defaultTemplate = "{{.GUID}}"
