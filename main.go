package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// ---- Template data structures ----------------------------------------

// DictCell represents one cell in the dictionary-selection table.
type DictCell struct {
	Valid   bool
	Index   int
	Title   string
	Checked bool
}

type headerData struct {
	ThemeCSS string
	Header   template.HTML
	IndexURL string
	Query    string
	Mode     string
	MaxHit   int
	DictRows [][]DictCell
}

type footerData struct {
	Footer template.HTML
}

type searchData struct {
	Results       []SearchResult
	SpellResults  []spellResult
	SectionAnchor template.HTML
}

type spellResult struct {
	Dict  string
	Words []string
}

type menuData struct {
	Entries []menuEntry
}

type menuEntry struct {
	DictTitle string
	Content   template.HTML
}

type referenceData struct {
	DictTitle string
	Item      template.HTML
}

// ---- HTTP handler ----------------------------------------------------

type appHandler struct {
	app *App
}

func (h *appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := ParseParams(r)

	// Binary/media modes: respond with raw bytes, no HTML wrapper.
	switch p.Mode {
	case "gaiji_w":
		data, err := h.app.WideGaiji(p.Book, p.Code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
		return

	case "gaiji_n":
		data, err := h.app.NarrowGaiji(p.Book, p.Code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
		return

	case "mono_graphic":
		data, err := h.app.ReadMonoGraphic(p.Book, p.Page, p.Offset, p.Width, p.Height)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/bmp")
		w.Write(data)
		return

	case "bmp":
		data, err := h.app.ReadColorGraphic(p.Book, p.Page, p.Offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/bmp")
		w.Write(data)
		return

	case "jpeg":
		data, err := h.app.ReadColorGraphic(p.Book, p.Page, p.Offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(data)
		return

	case "wave":
		data, err := h.app.ReadWave(p.Book, p.Page, p.Offset, p.Page2, p.Offset2)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "audio/wav")
		w.Write(data)
		return

	case "mpeg":
		data, err := h.app.ReadMPEG(p.Book, p.Page, p.Offset, p.Page2, p.Offset2)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "video/mpeg")
		w.Write(data)
		return
	}

	// All remaining modes produce an HTML page: header + body + footer.
	var buf bytes.Buffer
	h.renderHeader(&buf, p)

	switch {
	case p.Mode == "menu" || p.Mode == "copyright":
		h.renderMenu(&buf, p)
	case p.Mode == "reference":
		h.renderReference(&buf, p)
	case isSearchMode(p.Mode) || p.Query != "":
		h.renderSearch(&buf, p)
	default:
		if err := tHelp.Execute(&buf, nil); err != nil {
			log.Printf("tHelp: %v", err)
		}
	}

	h.renderFooter(&buf)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Write(buf.Bytes())
}

func isSearchMode(m string) bool {
	switch m {
	case "search", "exactsearch", "endsearch", "keywordsearch":
		return true
	}
	return false
}

// renderHeader writes the page header including the dict-selection table.
func (h *appHandler) renderHeader(buf *bytes.Buffer, p Params) {
	cfg := h.app.cfg
	cols := cfg.NumColumns
	if cols <= 0 {
		cols = 3
	}
	n := h.app.DictCount()

	var rows [][]DictCell
	for i := 0; i < n; i += cols {
		row := make([]DictCell, cols)
		for j := 0; j < cols; j++ {
			idx := i + j
			if idx < n {
				row[j] = DictCell{
					Valid:   true,
					Index:   idx,
					Title:   h.app.DictTitle(idx),
					Checked: isDictSelected(idx, p.Dict),
				}
			}
		}
		rows = append(rows, row)
	}

	d := headerData{
		ThemeCSS: cfg.ThemeCSS,
		Header:   cfg.Header,
		IndexURL: cfg.IndexURL,
		Query:    p.Query,
		Mode:     p.Mode,
		MaxHit:   p.MaxHit,
		DictRows: rows,
	}
	if err := tHeader.Execute(buf, d); err != nil {
		log.Printf("tHeader: %v", err)
	}
}

// isDictSelected returns true when idx is in the selection, or no selection exists.
func isDictSelected(idx int, selected []int) bool {
	if len(selected) == 0 {
		return true // all checked by default
	}
	for _, s := range selected {
		if s == idx {
			return true
		}
	}
	return false
}

func (h *appHandler) renderFooter(buf *bytes.Buffer) {
	if err := tFooter.Execute(buf, footerData{Footer: h.app.cfg.Footer}); err != nil {
		log.Printf("tFooter: %v", err)
	}
}

func (h *appHandler) renderSearch(buf *bytes.Buffer, p Params) {
	results := h.app.Search(p)

	var spellResults []spellResult
	for _, dict := range h.app.cfg.IspellDicts {
		if words := h.app.SpellCheck(p.Query, dict); len(words) > 0 {
			spellResults = append(spellResults, spellResult{Dict: dict, Words: words})
		}
	}

	d := searchData{
		Results:       results,
		SpellResults:  spellResults,
		SectionAnchor: h.app.cfg.SectionAnchor,
	}
	if err := tSearch.Execute(buf, d); err != nil {
		log.Printf("tSearch: %v", err)
	}
}

func (h *appHandler) renderMenu(buf *bytes.Buffer, p Params) {
	indices := p.Dict
	if len(indices) == 0 {
		for i := 0; i < h.app.DictCount(); i++ {
			indices = append(indices, i)
		}
	}

	var entries []menuEntry
	for _, i := range indices {
		var content template.HTML
		switch p.Mode {
		case "menu":
			content = h.app.MenuOf(i, p.Dict)
		case "copyright":
			content = h.app.CopyrightOf(i, p.Dict)
		}
		if content != "" {
			entries = append(entries, menuEntry{
				DictTitle: h.app.DictTitle(i),
				Content:   content,
			})
		}
	}

	if err := tMenu.Execute(buf, menuData{Entries: entries}); err != nil {
		log.Printf("tMenu: %v", err)
	}
}

func (h *appHandler) renderReference(buf *bytes.Buffer, p Params) {
	d := referenceData{
		DictTitle: h.app.DictTitle(p.Book),
		Item:      h.app.ContentAt(p.Book, p.Page, p.Offset, p.Dict),
	}
	if err := tReference.Execute(buf, d); err != nil {
		log.Printf("tReference: %v", err)
	}
}

// ---- Binary media delegators on App ---------------------------------
// These thin wrappers validate the book index before calling Subbook methods.

func (a *App) ReadColorGraphic(bookIdx, page, offset int) ([]byte, error) {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return nil, fmt.Errorf("invalid book index %d", bookIdx)
	}
	return a.dicts[bookIdx].ReadColorGraphic(page, offset)
}

func (a *App) ReadMonoGraphic(bookIdx, page, offset, width, height int) ([]byte, error) {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return nil, fmt.Errorf("invalid book index %d", bookIdx)
	}
	return a.dicts[bookIdx].ReadMonoGraphic(page, offset, width, height)
}

func (a *App) ReadWave(bookIdx, page, offset, page2, offset2 int) ([]byte, error) {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return nil, fmt.Errorf("invalid book index %d", bookIdx)
	}
	return a.dicts[bookIdx].ReadWave(page, offset, page2, offset2)
}

func (a *App) ReadMPEG(bookIdx, page, offset, page2, offset2 int) ([]byte, error) {
	if bookIdx < 0 || bookIdx >= len(a.dicts) {
		return nil, fmt.Errorf("invalid book index %d", bookIdx)
	}
	return a.dicts[bookIdx].ReadMPEG(page, offset, page2, offset2)
}

// ---- Entry point ----------------------------------------------------

func main() {
	// Explicitly register MIME types that may be missing or wrong in the
	// host OS database.  Without this, FileServer can serve .css as
	// text/plain, which browsers block when X-Content-Type-Options: nosniff
	// is in effect.
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")

	configPath := flag.String("config", "letmesee.json", "path to JSON config file")
	listen := flag.String("listen", ":8080", "address to listen on (host:port)")
	flag.Parse()

	if err := EBInit(); err != nil {
		log.Fatalf("EB init: %v", err)
	}
	defer EBFinalize()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("config file %q not found; starting with no dictionaries", *configPath)
			cfg = DefaultConfig()
		} else {
			log.Fatalf("load config: %v", err)
		}
	}

	app, err := NewApp(cfg)
	if err != nil {
		log.Fatalf("open dictionaries: %v", err)
	}
	defer app.Close()

	mux := http.NewServeMux()
	mux.Handle("/", &appHandler{app: app})
	// Serve theme CSS and assets from ./theme/ on disk.
	mux.Handle("/theme/", http.StripPrefix("/theme/", http.FileServer(http.Dir("theme"))))

	log.Printf("listening on %s", *listen)

	// Clean shutdown on SIGINT / SIGTERM.
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		log.Println("shutting down")
		app.Close()
		os.Exit(0)
	}()

	if err := http.ListenAndServe(*listen, mux); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
