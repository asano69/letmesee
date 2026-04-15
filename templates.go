package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
)

// staticTemplatesDir is the directory that contains all HTML templates.
// Templates are loaded once at startup; edit the files and restart to apply changes.
const staticTemplatesDir = "static/templates"

// mustLoadTemplate reads a template from staticTemplatesDir and fatally exits
// if the file is missing or cannot be parsed.
func mustLoadTemplate(name string) *template.Template {
	path := filepath.Join(staticTemplatesDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("template %s: %v", path, err)
	}
	return template.Must(template.New(name).Parse(string(data)))
}

// Parsed template instances, initialised once at startup.
var (
	tHeader    = mustLoadTemplate("header.html")
	tFooter    = mustLoadTemplate("footer.html")
	tHelp      = mustLoadTemplate("help.html")
	tSearch    = mustLoadTemplate("search.html")
	tMenu      = mustLoadTemplate("menu.html")
	tReference = mustLoadTemplate("reference.html")
)
