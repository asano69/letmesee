package main

/*
#cgo CFLAGS:  -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -leb

#include "hooks.h"
#include <eb/eb.h>
#include <eb/error.h>
#include <eb/text.h>
#include <eb/appendix.h>
#include <eb/font.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"unsafe"
)

// ------------------------------------------------------------------ //
// Library init / finalize                                              //
// ------------------------------------------------------------------ //

// EBInit initialises the EB library. Call once at program start.
func EBInit() error {
	if rc := C.eb_initialize_library(); rc != C.EB_SUCCESS {
		return fmt.Errorf("eb_initialize_library: %s", ebErrStr(rc))
	}
	return nil
}

// EBFinalize shuts the EB library down. Call at program exit.
func EBFinalize() { C.eb_finalize_library() }

// ------------------------------------------------------------------ //
// HookContext – opaque wrapper hiding CGo from other .go files         //
// ------------------------------------------------------------------ //

// HookContext wraps the C-heap EBHookContext.
// Callers must call Free() when done.
type HookContext struct {
	c *C.EBHookContext
}

// NewHookContext allocates and fills an EBHookContext on the C heap.
func NewHookContext(bookIdx int, indexURL string, forceInline bool,
	fontsize, fontsizeN, fontsizeW int, dictParams string) *HookContext {

	ctx := (*C.EBHookContext)(C.malloc(C.sizeof_EBHookContext))
	ctx.book_index = C.int(bookIdx)
	ctx.index_url = C.CString(indexURL)
	if forceInline {
		ctx.force_inline = 1
	} else {
		ctx.force_inline = 0
	}
	ctx.fontsize = C.int(fontsize)
	ctx.fontsize_n = C.int(fontsizeN)
	ctx.fontsize_w = C.int(fontsizeW)
	ctx.decoration = 0

	cDP := C.CString(dictParams)
	C.strncpy(&ctx.dict_params[0], cDP, 1023)
	C.free(unsafe.Pointer(cDP))

	return &HookContext{c: ctx}
}

// SetBookIndex updates the book_index field in place, so the same
// HookContext can be reused across subbooks without reallocation.
func (hc *HookContext) SetBookIndex(i int) {
	hc.c.book_index = C.int(i)
}

// Free releases the C-heap allocation.
func (hc *HookContext) Free() {
	C.free(unsafe.Pointer(hc.c.index_url))
	C.free(unsafe.Pointer(hc.c))
}

// ------------------------------------------------------------------ //
// Subbook – one dictionary volume inside an EPWING book                //
// ------------------------------------------------------------------ //

// Subbook wraps a bound EB_Book selected to a specific subbook,
// together with an optional appendix and a pre-built hookset.
type Subbook struct {
	book     *C.EB_Book
	appendix *C.EB_Appendix
	hookset  *C.EB_Hookset
	title    string // raw EUC-JP bytes
}

// Title returns the raw EUC-JP title string.
func (s *Subbook) Title() string { return s.title }

// Close releases all C-heap resources.
func (s *Subbook) Close() {
	C.eb_finalize_hookset(s.hookset)
	C.free(unsafe.Pointer(s.hookset))
	if s.appendix != nil {
		C.eb_finalize_appendix(s.appendix)
		C.free(unsafe.Pointer(s.appendix))
	}
	C.eb_finalize_book(s.book)
	C.free(unsafe.Pointer(s.book))
}

// ------------------------------------------------------------------ //
// Hit – one search result                                              //
// ------------------------------------------------------------------ //

// Hit holds one raw search result (EUC-JP bytes).
type Hit struct {
	Heading string
	Text    string
}

// ------------------------------------------------------------------ //
// OpenDictionaries                                                      //
// ------------------------------------------------------------------ //

// OpenDictionaries opens every subbook at dictPath (with optional
// appendixPath) and returns one Subbook per subbook found.
func OpenDictionaries(dictPath, appendixPath string) ([]*Subbook, error) {
	cPath := C.CString(dictPath)
	defer C.free(unsafe.Pointer(cPath))

	// A temporary master book is used only to list subbook codes.
	master := (*C.EB_Book)(C.malloc(C.sizeof_EB_Book))
	C.eb_initialize_book(master)
	defer func() {
		C.eb_finalize_book(master)
		C.free(unsafe.Pointer(master))
	}()

	if rc := C.eb_bind(master, cPath); rc != C.EB_SUCCESS {
		return nil, fmt.Errorf("eb_bind %q: %s", dictPath, ebErrStr(rc))
	}

	var codes [C.EB_MAX_SUBBOOKS]C.EB_Subbook_Code
	var count C.int
	if rc := C.eb_subbook_list(master, &codes[0], &count); rc != C.EB_SUCCESS {
		return nil, fmt.Errorf("eb_subbook_list %q: %s", dictPath, ebErrStr(rc))
	}

	var out []*Subbook
	for i := 0; i < int(count); i++ {
		s, err := openOneSubbook(dictPath, appendixPath, codes[i])
		if err != nil {
			continue // skip inaccessible subbooks
		}
		out = append(out, s)
	}
	return out, nil
}

func openOneSubbook(dictPath, appendixPath string, code C.EB_Subbook_Code) (*Subbook, error) {
	cPath := C.CString(dictPath)
	defer C.free(unsafe.Pointer(cPath))

	book := (*C.EB_Book)(C.malloc(C.sizeof_EB_Book))
	C.eb_initialize_book(book)

	if rc := C.eb_bind(book, cPath); rc != C.EB_SUCCESS {
		C.eb_finalize_book(book)
		C.free(unsafe.Pointer(book))
		return nil, fmt.Errorf("eb_bind: %s", ebErrStr(rc))
	}
	if rc := C.eb_set_subbook(book, code); rc != C.EB_SUCCESS {
		C.eb_finalize_book(book)
		C.free(unsafe.Pointer(book))
		return nil, fmt.Errorf("eb_set_subbook: %s", ebErrStr(rc))
	}

	var titleBuf [C.EB_MAX_TITLE_LENGTH + 1]C.char
	C.eb_subbook_title(book, &titleBuf[0])

	// Appendix is optional; failures are silently ignored.
	var app *C.EB_Appendix
	if appendixPath != "" {
		app = openAppendix(appendixPath)
	}

	hookset := (*C.EB_Hookset)(C.malloc(C.sizeof_EB_Hookset))
	C.eb_initialize_hookset(hookset)
	C.register_all_hooks(hookset)

	return &Subbook{
		book:     book,
		appendix: app,
		hookset:  hookset,
		title:    C.GoString(&titleBuf[0]),
	}, nil
}

func openAppendix(path string) *C.EB_Appendix {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	app := (*C.EB_Appendix)(C.malloc(C.sizeof_EB_Appendix))
	C.eb_initialize_appendix(app)

	if C.eb_bind_appendix(app, cPath) != C.EB_SUCCESS {
		C.eb_finalize_appendix(app)
		C.free(unsafe.Pointer(app))
		return nil
	}
	var codes [C.EB_MAX_SUBBOOKS]C.EB_Subbook_Code
	var count C.int
	if C.eb_appendix_subbook_list(app, &codes[0], &count) == C.EB_SUCCESS && count > 0 {
		C.eb_set_appendix_subbook(app, codes[0])
	}
	return app
}

// ------------------------------------------------------------------ //
// Search                                                               //
// ------------------------------------------------------------------ //

// Search runs the given mode query against this subbook and returns up
// to maxHit de-duplicated results.
// queryEUC must already be in EUC-JP encoding.
func (s *Subbook) Search(mode, queryEUC string, maxHit int, hc *HookContext) ([]Hit, error) {
	cQuery := C.CString(queryEUC)
	defer C.free(unsafe.Pointer(cQuery))

	var rc C.EB_Error_Code
	switch mode {
	case "exactsearch":
		rc = C.eb_search_exactword(s.book, cQuery)
	case "search":
		rc = C.eb_search_word(s.book, cQuery)
	case "endsearch":
		rc = C.eb_search_endword(s.book, cQuery)
	case "keywordsearch":
		rc = s.searchKeyword(queryEUC)
	default:
		rc = C.eb_search_exactword(s.book, cQuery)
	}
	if rc != C.EB_SUCCESS {
		return nil, nil
	}
	return s.collectHits(maxHit, hc)
}

// searchKeyword splits queryEUC on ASCII whitespace and EUC-JP
// ideographic space (\xa1\xa1), then calls eb_search_keyword.
func (s *Subbook) searchKeyword(queryEUC string) C.EB_Error_Code {
	parts := strings.FieldsFunc(queryEUC, func(r rune) bool {
		return r == ' ' || r == '\t'
	})
	var words []string
	for _, p := range parts {
		for _, w := range strings.Split(p, "\xa1\xa1") {
			if w != "" {
				words = append(words, w)
			}
		}
	}
	if len(words) == 0 {
		return C.EB_SUCCESS
	}
	cWords := make([]*C.char, len(words)+1)
	for i, w := range words {
		cWords[i] = C.CString(w)
	}
	cWords[len(words)] = nil
	defer func() {
		for i := range words {
			C.free(unsafe.Pointer(cWords[i]))
		}
	}()
	return C.eb_search_keyword(s.book,
		(**C.char)(unsafe.Pointer(&cWords[0])),
		C.int(len(words)))
}

func (s *Subbook) collectHits(maxHit int, hc *HookContext) ([]Hit, error) {
	hits := make([]C.EB_Hit, maxHit)
	var hitCount C.int
	if rc := C.eb_hit_list(s.book, C.int(maxHit), &hits[0], &hitCount); rc != C.EB_SUCCESS {
		return nil, nil
	}

	out := make([]Hit, 0, int(hitCount))
	seen := map[string]bool{}
	ctx := unsafe.Pointer(hc.c)

	for i := 0; i < int(hitCount); i++ {
		h := hits[i]

		C.eb_seek_text(s.book, &h.heading)
		var hLen C.size_t
		hPtr := C.read_heading_once(s.book, s.appendix, s.hookset, ctx, &hLen)
		heading := ""
		if hPtr != nil {
			heading = C.GoStringN(hPtr, C.int(hLen))
			C.free(unsafe.Pointer(hPtr))
		}

		C.eb_seek_text(s.book, &h.text)
		var tLen C.size_t
		tPtr := C.read_text_full(s.book, s.appendix, s.hookset, ctx, &tLen)
		text := ""
		if tPtr != nil {
			text = C.GoStringN(tPtr, C.int(tLen))
			C.free(unsafe.Pointer(tPtr))
		}

		key := heading + "\x00" + text
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, Hit{Heading: heading, Text: text})
	}
	return out, nil
}

// ------------------------------------------------------------------ //
// Content / Menu / Copyright                                           //
// ------------------------------------------------------------------ //

// Content reads text at an arbitrary page/offset position.
func (s *Subbook) Content(page, offset int, hc *HookContext) (string, error) {
	pos := C.EB_Position{page: C.int(page), offset: C.int(offset)}
	if rc := C.eb_seek_text(s.book, &pos); rc != C.EB_SUCCESS {
		return "", fmt.Errorf("eb_seek_text: %s", ebErrStr(rc))
	}
	var l C.size_t
	p := C.read_text_full(s.book, s.appendix, s.hookset, unsafe.Pointer(hc.c), &l)
	if p == nil {
		return "", nil
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoStringN(p, C.int(l)), nil
}

// Menu reads the book's menu entry, if available.
func (s *Subbook) Menu(hc *HookContext) (string, error) {
	if C.eb_have_menu(s.book) == 0 {
		return "", nil
	}
	var pos C.EB_Position
	if rc := C.eb_menu(s.book, &pos); rc != C.EB_SUCCESS {
		return "", nil
	}
	return s.Content(int(pos.page), int(pos.offset), hc)
}

// Copyright reads the book's copyright entry, if available.
func (s *Subbook) Copyright(hc *HookContext) (string, error) {
	if C.eb_have_copyright(s.book) == 0 {
		return "", nil
	}
	var pos C.EB_Position
	if rc := C.eb_copyright(s.book, &pos); rc != C.EB_SUCCESS {
		return "", nil
	}
	return s.Content(int(pos.page), int(pos.offset), hc)
}

// ------------------------------------------------------------------ //
// Gaiji (custom glyph) images                                          //
// ------------------------------------------------------------------ //

// WideGaiji returns the wide gaiji glyph as a PNG image.
func (s *Subbook) WideGaiji(code, fontCode int) ([]byte, error) {
	C.eb_set_font(s.book, C.EB_Font_Code(fontCode))
	size := wideGlyphBytes(fontCode)
	buf := make([]byte, size)
	if rc := C.eb_wide_font_character_bitmap(s.book, C.int(code),
		(*C.char)(unsafe.Pointer(&buf[0]))); rc != C.EB_SUCCESS {
		return nil, fmt.Errorf("eb_wide_font_character_bitmap: %s", ebErrStr(rc))
	}
	px := fontCodePx(fontCode)
	return bitmapToPNG(buf, px, px)
}

// NarrowGaiji returns the narrow gaiji glyph as a PNG image.
func (s *Subbook) NarrowGaiji(code, fontCode int) ([]byte, error) {
	C.eb_set_font(s.book, C.EB_Font_Code(fontCode))
	size := narrowGlyphBytes(fontCode)
	buf := make([]byte, size)
	if rc := C.eb_narrow_font_character_bitmap(s.book, C.int(code),
		(*C.char)(unsafe.Pointer(&buf[0]))); rc != C.EB_SUCCESS {
		return nil, fmt.Errorf("eb_narrow_font_character_bitmap: %s", ebErrStr(rc))
	}
	h := fontCodePx(fontCode)
	w := h / 2
	return bitmapToPNG(buf, w, h)
}

func fontCodePx(fc int) int {
	switch fc {
	case int(C.EB_FONT_24):
		return 24
	case int(C.EB_FONT_30):
		return 30
	case int(C.EB_FONT_48):
		return 48
	default:
		return 16
	}
}

func wideGlyphBytes(fc int) int   { px := fontCodePx(fc); return px * px / 8 }
func narrowGlyphBytes(fc int) int { px := fontCodePx(fc); return px * (px / 2) / 8 }

// bitmapToPNG converts a 1-bit-per-pixel EB glyph bitmap to PNG.
func bitmapToPNG(bitmap []byte, w, h int) ([]byte, error) {
	palette := color.Palette{color.White, color.Black}
	img := image.NewPaletted(image.Rect(0, 0, w, h), palette)
	bpr := (w + 7) / 8
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := y*bpr + x/8
			bit := uint(7 - x%8)
			if bi < len(bitmap) && (bitmap[bi]>>bit)&1 == 1 {
				img.SetColorIndex(x, y, 1)
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ------------------------------------------------------------------ //
// EBFontCode converts a fontsize integer to the EB_Font_Code constant  //
// ------------------------------------------------------------------ //

// EBFontCode maps a pixel font size to the EB library constant.
func EBFontCode(fontsize int) int {
	switch fontsize {
	case 24:
		return int(C.EB_FONT_24)
	case 30:
		return int(C.EB_FONT_30)
	case 48:
		return int(C.EB_FONT_48)
	default:
		return int(C.EB_FONT_16)
	}
}

// ------------------------------------------------------------------ //
// Utility                                                              //
// ------------------------------------------------------------------ //

func ebErrStr(rc C.EB_Error_Code) string {
	return C.GoString(C.eb_error_string(rc))
}
