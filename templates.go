package main

import (
	"text/template"
	"time"
)

var defaultFuncMap = map[string]interface{}{
	"date": func(t time.Time) string {
		return t.Format("January 2, 2006")
	},
	"format": func(fmt string, t time.Time) string {
		return t.Format(fmt)
	},
}

var tmpl = template.New("output").
	Funcs(defaultFuncMap)

var cmdTmpl = template.New("cmd").
	Funcs(defaultFuncMap)

var defaultTemplate = "{{.GUID}}"
