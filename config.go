package main

import (
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// staticThemeDir is the directory under which named themes are stored.
// A theme named "default" resolves to static/theme/default/default.css.
const staticThemeDir = "static/theme"

// CollectionSpec maps a collection name to a list of member specifiers.
// Each member is a string of the form "FolderName:N" where N is the
// 1-based subbook number within the named folder (the first subbook is :1).
type CollectionSpec map[string][]string

// ParseCollectionMember splits "FolderName:N" into its two components.
// The returned subbookNum is 1-based.
func ParseCollectionMember(s string) (folderName string, subbookNum int, err error) {
	idx := strings.LastIndexByte(s, ':')
	if idx < 0 {
		return "", 0, fmt.Errorf("collection member %q: missing ':N' suffix", s)
	}
	n, parseErr := strconv.Atoi(s[idx+1:])
	if parseErr != nil || n < 1 {
		return "", 0, fmt.Errorf("collection member %q: subbook number must be a positive integer", s)
	}
	return s[:idx], n, nil
}

// ---- On-disk YAML structures -----------------------------------------

// yamlServerBlock is the "server:" section of config.yaml.
type yamlServerBlock struct {
	Root string `yaml:"root"`
}

// yamlDictEntry is an optional explicit dictionary entry in config.yaml.
// Explicit entries are useful when a dictionary lives outside the root, or
// when an appendix needs to be attached.
type yamlDictEntry struct {
	Path     string `yaml:"path"`
	Appendix string `yaml:"appendix,omitempty"`
}

// yamlFile is the complete on-disk representation of config.yaml.
type yamlFile struct {
	Server        yamlServerBlock     `yaml:"server"`
	DictList      []yamlDictEntry     `yaml:"dictlist,omitempty"`
	Collections   map[string][]string `yaml:"collections,omitempty"`
	NumColumns    int                 `yaml:"num_columns,omitempty"`
	IspellCommand string              `yaml:"ispell_command,omitempty"`
	IspellDicts   []string            `yaml:"ispell_dict_list,omitempty"`
	FontSize      int                 `yaml:"fontsize,omitempty"`
	// ForceInline is a pointer so that omitting the field in YAML preserves
	// the default value set by DefaultConfig rather than always forcing false.
	ForceInline   *bool  `yaml:"force_inline,omitempty"`
	IndexURL      string `yaml:"index,omitempty"`
	Header        string `yaml:"header,omitempty"`
	Footer        string `yaml:"footer,omitempty"`
	Theme         string `yaml:"theme,omitempty"`
	CSS           string `yaml:"css,omitempty"`
	SectionAnchor string `yaml:"section_anchor,omitempty"`
}

// LoadYAMLConfig reads a YAML config file and merges it over the defaults.
// Unset fields keep their default values.
func LoadYAMLConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	var yc yamlFile
	if err := yaml.NewDecoder(f).Decode(&yc); err != nil {
		return cfg, fmt.Errorf("decode %q: %w", path, err)
	}

	cfg.DictRoot = yc.Server.Root

	for _, d := range yc.DictList {
		cfg.DictList = append(cfg.DictList, DictEntry{
			Path:     d.Path,
			Appendix: d.Appendix,
		})
	}

	if len(yc.Collections) > 0 {
		cfg.Collections = CollectionSpec(yc.Collections)
	}

	if yc.NumColumns > 0 {
		cfg.NumColumns = yc.NumColumns
	}
	if yc.IspellCommand != "" {
		cfg.IspellCommand = yc.IspellCommand
	}
	if len(yc.IspellDicts) > 0 {
		cfg.IspellDicts = yc.IspellDicts
	}
	if yc.FontSize > 0 {
		cfg.FontSize = yc.FontSize
	}
	// Only override the default when force_inline is explicitly set in the file.
	if yc.ForceInline != nil {
		cfg.ForceInline = *yc.ForceInline
	}
	if yc.IndexURL != "" {
		cfg.IndexURL = yc.IndexURL
	}
	cfg.Header = template.HTML(yc.Header)
	cfg.Footer = template.HTML(yc.Footer)
	cfg.SectionAnchor = template.HTML(yc.SectionAnchor)

	// Theme takes precedence over a raw CSS URL.
	// CSS files for named themes live under static/theme/<name>/<name>.css.
	switch {
	case yc.Theme != "":
		cfg.ThemeCSS = staticThemeDir + "/" + yc.Theme + "/" + yc.Theme + ".css"
	case yc.CSS != "":
		cfg.ThemeCSS = yc.CSS
	}

	return cfg, nil
}
