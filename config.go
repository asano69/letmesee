package main

import (
	"encoding/json"
	"html/template"
	"os"
)

// JSONDictEntry mirrors one element of the "dictlist" JSON array.
// Appendix is optional.
type JSONDictEntry struct {
	Path     string `json:"path"`
	Appendix string `json:"appendix,omitempty"`
}

// JSONConfig is the on-disk format, analogous to letmesee.conf.
//
// Example file:
//
//	{
//	  "dictlist": [
//	    {"path": "/usr/share/dict/genius"},
//	    {"path": "/usr/share/dict/mydict", "appendix": "/usr/share/dict/mydict-appendix"}
//	  ],
//	  "num_columns": 3,
//	  "fontsize": 16,
//	  "theme": "default",
//	  "force_inline": true
//	}
type JSONConfig struct {
	DictList      []JSONDictEntry `json:"dictlist"`
	NumColumns    int             `json:"num_columns"`
	IspellCommand string          `json:"ispell_command"`
	IspellDicts   []string        `json:"ispell_dict_list"`
	FontSize      int             `json:"fontsize"`
	// ForceInline is a pointer so that omitting the field in JSON preserves
	// the default value set by DefaultConfig rather than always forcing false.
	ForceInline   *bool  `json:"force_inline,omitempty"`
	IndexURL      string `json:"index"`
	Header        string `json:"header"`
	Footer        string `json:"footer"`
	Theme         string `json:"theme"`
	CSS           string `json:"css"`
	SectionAnchor string `json:"section_anchor"`
}

// LoadConfig reads a JSON config file and merges it over the defaults.
// Unset fields in the file keep their default values.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	var jc JSONConfig
	if err := json.NewDecoder(f).Decode(&jc); err != nil {
		return cfg, err
	}

	for _, d := range jc.DictList {
		cfg.DictList = append(cfg.DictList, DictEntry{
			Path:     d.Path,
			Appendix: d.Appendix,
		})
	}
	if jc.NumColumns > 0 {
		cfg.NumColumns = jc.NumColumns
	}
	if jc.IspellCommand != "" {
		cfg.IspellCommand = jc.IspellCommand
	}
	if len(jc.IspellDicts) > 0 {
		cfg.IspellDicts = jc.IspellDicts
	}
	if jc.FontSize > 0 {
		cfg.FontSize = jc.FontSize
	}
	// Only override the default when force_inline is explicitly set in the file.
	if jc.ForceInline != nil {
		cfg.ForceInline = *jc.ForceInline
	}
	if jc.IndexURL != "" {
		cfg.IndexURL = jc.IndexURL
	}
	cfg.Header = template.HTML(jc.Header)
	cfg.Footer = template.HTML(jc.Footer)
	cfg.SectionAnchor = template.HTML(jc.SectionAnchor)

	// Theme takes precedence over raw CSS URL.
	switch {
	case jc.Theme != "":
		cfg.ThemeCSS = "theme/" + jc.Theme + "/" + jc.Theme + ".css"
	case jc.CSS != "":
		cfg.ThemeCSS = jc.CSS
	}

	return cfg, nil
}
