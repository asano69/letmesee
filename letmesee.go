package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ------------------------------------------------------------------ //
// Configuration                                                        //
// ------------------------------------------------------------------ //

// DictEntry describes one dictionary directory and an optional appendix.
type DictEntry struct {
	Path     string
	Appendix string
}

// Config holds values equivalent to letmesee.conf.
type Config struct {
	DictList      []DictEntry
	NumColumns    int
	IspellCommand string
	IspellDicts   []string
	FontSize      int // 16, 24, 30, or 48
	ForceInline   bool
	IndexURL      string
	Header        template.HTML
	Footer        template.HTML
	ThemeCSS      string // URL to CSS file
	SectionAnchor template.HTML
}

// DefaultConfig returns a Config with the same defaults as letmesee.rb,
// except that ForceInline defaults to true so that encyclopedia illustrations
// and tables are displayed inline rather than as follow-through links.
func DefaultConfig() Config {
	return Config{
		NumColumns:    3,
		IspellCommand: "ispell",
		IspellDicts:   []string{"american"},
		FontSize:      16,
		IndexURL:      "./",
		ForceInline:   true,
	}
}

// fontDimensions returns (size, narrowW, wideW, ebFontCode).
func (c *Config) fontDimensions() (size, narrowW, wideW, code int) {
	switch c.FontSize {
	case 24:
		return 24, 12, 24, EBFontCode(24)
	case 30:
		return 30, 16, 32, EBFontCode(30)
	case 48:
		return 48, 24, 48, EBFontCode(48)
	default:
		return 16, 8, 16, EBFontCode(16)
	}
}

// ------------------------------------------------------------------ //
// Application                                                          //
// ------------------------------------------------------------------ //

// App owns all open subbooks and the active config.
type App struct {
	cfg   Config
	dicts []*Subbook
}

// NewApp opens every dictionary listed in cfg.
func NewApp(cfg Config) (*App, error) {
	a := &App{cfg: cfg}
	for _, de := range cfg.DictList {
		subs, err := OpenDictionaries(de.Path, de.Appendix)
		if err != nil {
			return nil, fmt.Errorf("opening %q: %w", de.Path, err)
		}
		a.dicts = append(a.dicts, subs...)
	}
	return a, nil
}

// Close releases all open dictionaries.
func (a *App) Close() {
	for _, d := range a.dicts {
		d.Close()
	}
}

// DictCount returns the number of open subbooks.
func (a *App) DictCount() int { return len(a.dicts) }

// DictTitle returns the UTF-8 title for subbook i.
func (a *App) DictTitle(i int) string {
	if i < 0 || i >= len(a.dicts) {
		return ""
	}
	return eucToUTF8(a.dicts[i].Title())
}

// ------------------------------------------------------------------ //
// Request parameters                                                   //
// ------------------------------------------------------------------ //

// Params holds all parsed CGI query parameters.
type Params struct {
	Query   string
	Mode    string
	MaxHit  int
	Book    int
	Dict    []int  // selected subbook indices; empty means all
	IE      string // declared input encoding of Query
	Page    int
	Offset  int
	Page2   int
	Offset2 int
	Width   int
	Height  int
	Code    int
}

// ParseParams reads URL query values from r into a Params.
func ParseParams(r *http.Request) Params {
	q := r.URL.Query()
	p := Params{
		Query:   q.Get("query"),
		Mode:    q.Get("mode"),
		MaxHit:  10,
		IE:      q.Get("ie"),
		Page:    intQ(q, "page"),
		Offset:  intQ(q, "offset"),
		Page2:   intQ(q, "page2"),
		Offset2: intQ(q, "offset2"),
		Width:   intQ(q, "width"),
		Height:  intQ(q, "height"),
		Code:    intQ(q, "code"),
		Book:    intQ(q, "book"),
	}
	if mh := intQ(q, "maxhit"); mh > 0 {
		p.MaxHit = mh
	}
	if p.Mode == "" {
		p.Mode = "exactsearch"
	}
	if p.IE == "" {
		p.IE = "UTF-8"
	}
	if p.Query != "" {
		p.Query = normaliseInputEncoding(p.Query, p.IE)
	}
	for _, v := range q["dict"] {
		n := 0
		fmt.Sscanf(v, "%d", &n)
		p.Dict = append(p.Dict, n)
	}
	return p
}

func intQ(q url.Values, key string) int {
	n := 0
	fmt.Sscanf(q.Get(key), "%d", &n)
	return n
}

// normaliseInputEncoding converts the raw CGI query bytes to UTF-8.
func normaliseInputEncoding(s, ie string) string {
	switch strings.ToUpper(strings.ReplaceAll(ie, "-", "")) {
	case "EUCJP", "EUC":
		if u, err := eucToUTF8Err(s); err == nil {
			return u
		}
	}
	return s
}

// ------------------------------------------------------------------ //
// Search                                                               //
// ------------------------------------------------------------------ //

// SearchResult holds one dictionary's worth of rendered results.
type SearchResult struct {
	DictIndex int
	DictTitle string
	Hits      []HitHTML
}

// HitHTML holds HTML-safe heading and content for one hit.
type HitHTML struct {
	Heading template.HTML
	Content template.HTML
}

// Search queries the selected dictionaries and returns rendered results.
func (a *App) Search(p Params) []SearchResult {
	indices := p.Dict
	if len(indices) == 0 {
		for i := range a.dicts {
			indices = append(indices, i)
		}
	}

	var results []SearchResult
	for _, i := range indices {
		if i < 0 || i >= len(a.dicts) {
			continue
		}
		hc := a.makeHookContext(i, p.Dict)
		hits := a.trySearch(a.dicts[i], p, hc)
		hc.Free()
		if len(hits) == 0 {
			continue
		}
		sr := SearchResult{
			DictIndex: i,
			DictTitle: eucToUTF8(a.dicts[i].Title()),
		}
		for _, h := range hits {
			sr.Hits = append(sr.Hits, HitHTML{
				Heading: a.htmlOutput(eucToUTF8(h.Heading)),
				Content: a.htmlOutput(eucToUTF8(h.Text)),
			})
		}
		results = append(results, sr)
	}
	return results
}

// trySearch tries each stemmed variant of the query until a non-empty
// result is returned, mirroring the Ruby search loop.
func (a *App) trySearch(sub *Subbook, p Params, hc *HookContext) []Hit {
	// Stem works on UTF-8; apply it before converting to EUC-JP.
	variants := strings.Split(Stem(convertToASCII(p.Query)), "|")

	for _, variant := range variants {
		queryEUC, err := utf8ToEUC(variant)
		if err != nil {
			queryEUC = variant
		}
		hits, _ := sub.Search(p.Mode, queryEUC, p.MaxHit, hc)
		if len(hits) > 0 {
			return hits
		}
	}
	return nil
}

// ------------------------------------------------------------------ //
// Reference / Menu / Copyright                                         //
// ------------------------------------------------------------------ //

// ContentAt returns the HTML content at a page/offset reference.
func (a *App) ContentAt(bookIdx, page, offset int, dictSel []int) template.HTML {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return ""
	}
	hc := a.makeHookContext(bookIdx, dictSel)
	defer hc.Free()
	raw, _ := a.dicts[bookIdx].Content(page, offset, hc)
	return a.htmlOutput(eucToUTF8(raw))
}

// MenuOf returns the HTML menu for subbook bookIdx.
func (a *App) MenuOf(bookIdx int, dictSel []int) template.HTML {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return ""
	}
	hc := a.makeHookContext(bookIdx, dictSel)
	defer hc.Free()
	raw, _ := a.dicts[bookIdx].Menu(hc)
	return a.htmlOutput(eucToUTF8(raw))
}

// CopyrightOf returns the HTML copyright notice for subbook bookIdx.
func (a *App) CopyrightOf(bookIdx int, dictSel []int) template.HTML {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return ""
	}
	hc := a.makeHookContext(bookIdx, dictSel)
	defer hc.Free()
	raw, _ := a.dicts[bookIdx].Copyright(hc)
	return a.htmlOutput(eucToUTF8(raw))
}

// ------------------------------------------------------------------ //
// Gaiji / binary media access                                          //
// ------------------------------------------------------------------ //

// WideGaiji returns the wide gaiji glyph for bookIdx as PNG bytes.
func (a *App) WideGaiji(bookIdx, code int) ([]byte, error) {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return nil, fmt.Errorf("invalid book index %d", bookIdx)
	}
	_, _, _, fc := a.cfg.fontDimensions()
	return a.dicts[bookIdx].WideGaiji(code, fc)
}

// NarrowGaiji returns the narrow gaiji glyph for bookIdx as PNG bytes.
func (a *App) NarrowGaiji(bookIdx, code int) ([]byte, error) {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return nil, fmt.Errorf("invalid book index %d", bookIdx)
	}
	_, _, _, fc := a.cfg.fontDimensions()
	return a.dicts[bookIdx].NarrowGaiji(code, fc)
}

// ------------------------------------------------------------------ //
// Spell checking via ispell                                            //
// ------------------------------------------------------------------ //

// SpellCheck returns spelling suggestions using the named ispell dictionary.
func (a *App) SpellCheck(word, dict string) []string {
	cmd := exec.Command(a.cfg.IspellCommand, "-a", "-m", "-C", "-d", dict)
	cmd.Stdin = strings.NewReader(toISO8859(word) + "\n")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseIspellOutput(string(out))
}

func parseIspellOutput(out string) []string {
	lines := strings.SplitN(out, "\n", 3)
	if len(lines) < 2 {
		return nil
	}
	line := lines[1] // skip the version header
	if strings.HasPrefix(line, "+ ") {
		return []string{strings.ToLower(strings.TrimPrefix(line, "+ "))}
	}
	if strings.HasPrefix(line, "& ") {
		if idx := strings.Index(line, ": "); idx >= 0 {
			return strings.Split(line[idx+2:], ", ")
		}
	}
	return nil
}

// ------------------------------------------------------------------ //
// HookContext construction helpers                                     //
// ------------------------------------------------------------------ //

func (a *App) makeHookContext(bookIdx int, dictSel []int) *HookContext {
	size, narrow, wide, _ := a.cfg.fontDimensions()
	return NewHookContext(bookIdx, a.cfg.IndexURL, a.cfg.ForceInline,
		size, narrow, wide, buildDictParams(dictSel))
}

func buildDictParams(dicts []int) string {
	var b strings.Builder
	for _, d := range dicts {
		fmt.Fprintf(&b, "&dict=%d", d)
	}
	return b.String()
}

// ------------------------------------------------------------------ //
// HTML post-processing                                                 //
// ------------------------------------------------------------------ //

var (
	reRef     = regexp.MustCompile(`<reference>(.*?)</reference ([^>]+)>`)
	reMonoGfx = regexp.MustCompile(`<mono_graphic ([^>]+)>(.*?)</mono_graphic ([^>]+)>`)
	// reDecoTag matches <deco_TAG> and </deco_TAG> pairs emitted by hooks.
	// The Ruby port changed emphasis hooks to use <deco_strong>, <deco_i>,
	// <deco_b>, <deco_sub>, <deco_sup> so that html_output can expand them.
	reDecoOpen  = regexp.MustCompile(`<deco_([a-z]+)>`)
	reDecoClose = regexp.MustCompile(`</deco_([a-z]+)>`)
)

// htmlOutput mirrors the Ruby html_output method: it unescapes the
// \< / \> markers written by hook callbacks and expands the special
// <reference>, <mono_graphic>, and <deco_X> pseudo-tags into real HTML.
func (a *App) htmlOutput(s string) template.HTML {
	s = strings.ReplaceAll(s, `\<`, "<")
	s = strings.ReplaceAll(s, `\>`, ">")
	s = strings.ReplaceAll(s, `\"`, `"`)

	s = reRef.ReplaceAllStringFunc(s, func(m string) string {
		g := reRef.FindStringSubmatch(m)
		if len(g) < 3 {
			return m
		}
		inner, attrs := g[1], g[2]
		return fmt.Sprintf(`<a href="%s?mode=reference&amp;%s">%s</a>`,
			a.cfg.IndexURL, attrs, inner)
	})

	s = reMonoGfx.ReplaceAllStringFunc(s, func(m string) string {
		g := reMonoGfx.FindStringSubmatch(m)
		if len(g) < 4 {
			return m
		}
		dimAttrs, alt, posAttrs := g[1], g[2], g[3]
		return fmt.Sprintf(
			`<img src="%s?mode=mono_graphic&amp;%s&amp;%s" alt="%s">`,
			a.cfg.IndexURL, posAttrs, dimAttrs, alt)
	})

	// Expand <deco_TAG> / </deco_TAG> into real HTML elements.
	// Unknown tag names pass through harmlessly because browsers ignore
	// unrecognised tags; only well-known inline elements are emitted.
	s = reDecoOpen.ReplaceAllString(s, "<$1>")
	s = reDecoClose.ReplaceAllString(s, "</$1>")

	return template.HTML(s)
}

// ------------------------------------------------------------------ //
// Encoding helpers                                                     //
// ------------------------------------------------------------------ //

func eucToUTF8(s string) string {
	out, err := eucToUTF8Err(s)
	if err != nil {
		return s
	}
	return out
}

func eucToUTF8Err(s string) (string, error) {
	b, _, err := transform.Bytes(japanese.EUCJP.NewDecoder(), []byte(s))
	return string(b), err
}

func utf8ToEUC(s string) (string, error) {
	b, _, err := transform.Bytes(japanese.EUCJP.NewEncoder(), []byte(s))
	return string(b), err
}

// convertToASCII strips Latin diacritics, matching the Ruby method of the
// same name.
func convertToASCII(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue // strip combining character
		}
		b.WriteRune(r)
	}
	return b.String()
}

// toISO8859 drops any character outside the Latin-1 range, producing a
// safe string to hand to ispell.
func toISO8859(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r <= 0xFF {
			b.WriteRune(r)
		}
	}
	return b.String()
}
