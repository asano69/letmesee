#include "hooks.h"
#include <eb/binary.h>

/* ------------------------------------------------------------------ */
/* Hook callbacks                                                       */
/* ------------------------------------------------------------------ */

static EB_Error_Code
hook_newline(EB_Book *book, EB_Appendix *app, void *container,
             EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\<br\\>\n");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_wide_font(EB_Book *book, EB_Appendix *app, void *container,
               EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[512];
    snprintf(buf, sizeof(buf),
             "\\<img src=\"%s?book=%d&mode=gaiji_w&code=%u\""
             " alt=\"_\" width=\"%d\" height=\"%d\"\\>",
             ctx->index_url, ctx->book_index, argv[0],
             ctx->fontsize_w, ctx->fontsize);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_narrow_font(EB_Book *book, EB_Appendix *app, void *container,
                 EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[512];
    snprintf(buf, sizeof(buf),
             "\\<img src=\"%s?book=%d&mode=gaiji_n&code=%u\""
             " alt=\"_\" width=\"%d\" height=\"%d\"\\>",
             ctx->index_url, ctx->book_index, argv[0],
             ctx->fontsize_n, ctx->fontsize);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_emphasis(EB_Book *book, EB_Appendix *app, void *container,
                    EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\<strong\\>");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_emphasis(EB_Book *book, EB_Appendix *app, void *container,
                  EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\</strong\\>");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_subscript(EB_Book *book, EB_Appendix *app, void *container,
                     EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\<sub\\>");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_subscript(EB_Book *book, EB_Appendix *app, void *container,
                   EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\</sub\\>");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_superscript(EB_Book *book, EB_Appendix *app, void *container,
                       EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\<sup\\>");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_superscript(EB_Book *book, EB_Appendix *app, void *container,
                     EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\</sup\\>");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_reference(EB_Book *book, EB_Appendix *app, void *container,
                     EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\<span class=\"reference\"\\>\\<reference\\>");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_reference(EB_Book *book, EB_Appendix *app, void *container,
                   EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[2048];
    snprintf(buf, sizeof(buf),
             "\\</reference book=%d&page=%u&offset=%u%s\\>\\</span\\>",
             ctx->book_index, argv[1], argv[2], ctx->dict_params);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_candidate(EB_Book *book, EB_Appendix *app, void *container,
                     EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\<span class=\"reference\"\\>\\<reference\\>");
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_candidate_group(EB_Book *book, EB_Appendix *app, void *container,
                         EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[2048];
    snprintf(buf, sizeof(buf),
             "\\</reference book=%d&page=%u&offset=%u%s\\>\\</span\\>",
             ctx->book_index, argv[1], argv[2], ctx->dict_params);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_mono_graphic(EB_Book *book, EB_Appendix *app, void *container,
                        EB_Hook_Code code, int argc, const unsigned int *argv)
{
    char buf[256];
    /* argv[3] = width, argv[2] = height */
    snprintf(buf, sizeof(buf),
             "\\<mono_graphic width=%u&height=%u\\>", argv[3], argv[2]);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_mono_graphic(EB_Book *book, EB_Appendix *app, void *container,
                      EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[256];
    snprintf(buf, sizeof(buf),
             "\\</mono_graphic book=%d&page=%u&offset=%u\\>",
             ctx->book_index, argv[1], argv[2]);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_color_bmp(EB_Book *book, EB_Appendix *app, void *container,
                     EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[512];
    if (ctx->force_inline) {
        snprintf(buf, sizeof(buf),
                 "\\<img src=\"%s?mode=bmp&book=%d&page=%u&offset=%u\""
                 " alt=\"[image]\"\\>",
                 ctx->index_url, ctx->book_index, argv[2], argv[3]);
    } else {
        snprintf(buf, sizeof(buf),
                 "\\<a href=\"%s?mode=bmp&book=%d&page=%u&offset=%u\"\\>[image] ",
                 ctx->index_url, ctx->book_index, argv[2], argv[3]);
    }
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_color_jpeg(EB_Book *book, EB_Appendix *app, void *container,
                      EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[512];
    if (ctx->force_inline) {
        snprintf(buf, sizeof(buf),
                 "\\<img src=\"%s?mode=jpeg&book=%d&page=%u&offset=%u\""
                 " alt=\"[image]\"\\>",
                 ctx->index_url, ctx->book_index, argv[2], argv[3]);
    } else {
        snprintf(buf, sizeof(buf),
                 "\\<a href=\"%s?mode=jpeg&book=%d&page=%u&offset=%u\"\\>[image] ",
                 ctx->index_url, ctx->book_index, argv[2], argv[3]);
    }
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_color_graphic(EB_Book *book, EB_Appendix *app, void *container,
                       EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    if (!ctx->force_inline) {
        eb_write_text_string(book, "\\</a\\>");
    }
    return EB_SUCCESS;
}

#ifdef EB_HOOK_BEGIN_IN_COLOR_BMP
static EB_Error_Code
hook_begin_in_color_bmp(EB_Book *book, EB_Appendix *app, void *container,
                        EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[512];
    snprintf(buf, sizeof(buf),
             "\\<img src=\"%s?mode=bmp&book=%d&page=%u&offset=%u\""
             " alt=\"[image]\"\\>",
             ctx->index_url, ctx->book_index, argv[2], argv[3]);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_in_color_jpeg(EB_Book *book, EB_Appendix *app, void *container,
                         EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[512];
    snprintf(buf, sizeof(buf),
             "\\<img src=\"%s?mode=jpeg&book=%d&page=%u&offset=%u\""
             " alt=\"[image]\"\\>",
             ctx->index_url, ctx->book_index, argv[2], argv[3]);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}
#endif /* EB_HOOK_BEGIN_IN_COLOR_BMP */

/*
 * Audio: emit an HTML5 <audio> element directly so the browser can play
 * the sound inline without navigating away.  The full element is written
 * in the begin hook; the end hook writes nothing because all required
 * information (page/offset of both endpoints) is available in argv here.
 */
static EB_Error_Code
hook_begin_wave(EB_Book *book, EB_Appendix *app, void *container,
                EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[512];
    snprintf(buf, sizeof(buf),
             "\\<audio controls src=\"%s?mode=wave&book=%d"
             "&page=%u&offset=%u&page2=%u&offset2=%u\"\\>\\</audio\\>",
             ctx->index_url, ctx->book_index,
             argv[2], argv[3], argv[4], argv[5]);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_wave(EB_Book *book, EB_Appendix *app, void *container,
              EB_Hook_Code code, int argc, const unsigned int *argv)
{
    /* Audio element is complete from hook_begin_wave; nothing to add. */
    return EB_SUCCESS;
}

static EB_Error_Code
hook_begin_mpeg(EB_Book *book, EB_Appendix *app, void *container,
                EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    char buf[512];
    snprintf(buf, sizeof(buf),
             "\\<a href=\"%s?mode=mpeg&book=%d"
             "&page=%u&offset=%u&page2=%u&offset2=%u\"\\>[video] ",
             ctx->index_url, ctx->book_index,
             argv[2], argv[3], argv[4], argv[5]);
    eb_write_text_string(book, buf);
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_mpeg(EB_Book *book, EB_Appendix *app, void *container,
              EB_Hook_Code code, int argc, const unsigned int *argv)
{
    eb_write_text_string(book, "\\</a\\>");
    return EB_SUCCESS;
}

#ifdef EB_HOOK_BEGIN_DECORATION
static EB_Error_Code
hook_begin_decoration(EB_Book *book, EB_Appendix *app, void *container,
                      EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    ctx->decoration = (int)argv[1];
    switch (ctx->decoration) {
    case 1: eb_write_text_string(book, "\\<i\\>"); break;
    case 3: eb_write_text_string(book, "\\<b\\>"); break;
    }
    return EB_SUCCESS;
}

static EB_Error_Code
hook_end_decoration(EB_Book *book, EB_Appendix *app, void *container,
                    EB_Hook_Code code, int argc, const unsigned int *argv)
{
    EBHookContext *ctx = (EBHookContext *)container;
    switch (ctx->decoration) {
    case 1: eb_write_text_string(book, "\\</i\\>"); break;
    case 3: eb_write_text_string(book, "\\</b\\>"); break;
    }
    return EB_SUCCESS;
}
#endif /* EB_HOOK_BEGIN_DECORATION */

/* ------------------------------------------------------------------ */
/* Hook registration                                                    */
/* ------------------------------------------------------------------ */

void
register_all_hooks(EB_Hookset *hookset)
{
    static const EB_Hook hooks[] = {
        {EB_HOOK_NEWLINE,              hook_newline},
        {EB_HOOK_WIDE_FONT,            hook_wide_font},
        {EB_HOOK_NARROW_FONT,          hook_narrow_font},
        {EB_HOOK_BEGIN_EMPHASIS,       hook_begin_emphasis},
        {EB_HOOK_END_EMPHASIS,         hook_end_emphasis},
        {EB_HOOK_BEGIN_SUBSCRIPT,      hook_begin_subscript},
        {EB_HOOK_END_SUBSCRIPT,        hook_end_subscript},
        {EB_HOOK_BEGIN_SUPERSCRIPT,    hook_begin_superscript},
        {EB_HOOK_END_SUPERSCRIPT,      hook_end_superscript},
        {EB_HOOK_BEGIN_REFERENCE,      hook_begin_reference},
        {EB_HOOK_END_REFERENCE,        hook_end_reference},
        {EB_HOOK_BEGIN_CANDIDATE,      hook_begin_candidate},
        {EB_HOOK_END_CANDIDATE_GROUP,  hook_end_candidate_group},
        {EB_HOOK_BEGIN_MONO_GRAPHIC,   hook_begin_mono_graphic},
        {EB_HOOK_END_MONO_GRAPHIC,     hook_end_mono_graphic},
        {EB_HOOK_BEGIN_COLOR_BMP,      hook_begin_color_bmp},
        {EB_HOOK_BEGIN_COLOR_JPEG,     hook_begin_color_jpeg},
        {EB_HOOK_END_COLOR_GRAPHIC,    hook_end_color_graphic},
        {EB_HOOK_BEGIN_WAVE,           hook_begin_wave},
        {EB_HOOK_END_WAVE,             hook_end_wave},
        {EB_HOOK_BEGIN_MPEG,           hook_begin_mpeg},
        {EB_HOOK_END_MPEG,             hook_end_mpeg},
#ifdef EB_HOOK_BEGIN_IN_COLOR_BMP
        {EB_HOOK_BEGIN_IN_COLOR_BMP,   hook_begin_in_color_bmp},
        {EB_HOOK_BEGIN_IN_COLOR_JPEG,  hook_begin_in_color_jpeg},
#endif
#ifdef EB_HOOK_BEGIN_DECORATION
        {EB_HOOK_BEGIN_DECORATION,     hook_begin_decoration},
        {EB_HOOK_END_DECORATION,       hook_end_decoration},
#endif
        {EB_HOOK_NULL, NULL}
    };

    int i;
    for (i = 0; hooks[i].code != EB_HOOK_NULL; i++) {
        eb_set_hook(hookset, &hooks[i]);
    }
}

/* ------------------------------------------------------------------ */
/* Text reading helpers                                                 */
/* ------------------------------------------------------------------ */

#define READ_BUF 65536

char *
read_text_full(EB_Book *book, EB_Appendix *appendix,
               EB_Hookset *hookset, void *ctx, size_t *out_len)
{
    size_t total  = 0;
    size_t alloc  = READ_BUF;
    char  *result = (char *)malloc(alloc);
    char   buf[READ_BUF];
    ssize_t chunk;

    if (!result) { *out_len = 0; return NULL; }

    for (;;) {
        EB_Error_Code rc = eb_read_text(book, appendix, hookset, ctx,
                                        READ_BUF - 1, buf, &chunk);
        /* EB_ERR_END_OF_CONTENT signals normal end; any other error aborts. */
        if (rc == EB_ERR_END_OF_CONTENT) break;
        if (rc != EB_SUCCESS) break;
        /* Some EB versions signal end-of-content with SUCCESS + zero bytes. */
        if (chunk <= 0) break;
        if (total + (size_t)chunk + 1 >= alloc) {
            alloc = (total + (size_t)chunk) * 2 + 1;
            char *tmp = (char *)realloc(result, alloc);
            if (!tmp) { free(result); *out_len = 0; return NULL; }
            result = tmp;
        }
        memcpy(result + total, buf, (size_t)chunk);
        total += (size_t)chunk;
    }
    result[total] = '\0';
    *out_len = total;
    return result;
}

char *
read_heading_once(EB_Book *book, EB_Appendix *appendix,
                  EB_Hookset *hookset, void *ctx, size_t *out_len)
{
    char    buf[READ_BUF];
    ssize_t len;

    EB_Error_Code rc = eb_read_heading(book, appendix, hookset, ctx,
                                       READ_BUF - 1, buf, &len);
    if (rc != EB_SUCCESS) { *out_len = 0; return NULL; }

    char *result = (char *)malloc((size_t)len + 1);
    if (!result)  { *out_len = 0; return NULL; }
    memcpy(result, buf, (size_t)len);
    result[len] = '\0';
    *out_len = (size_t)len;
    return result;
}

/* ------------------------------------------------------------------ */
/* Binary data reading helper                                           */
/* ------------------------------------------------------------------ */

char *
read_binary_all(EB_Book *book, size_t *out_len)
{
    size_t total = 0;
    size_t alloc = READ_BUF;
    char  *result = (char *)malloc(alloc);
    char   buf[4096];
    ssize_t chunk;

    if (!result) { *out_len = 0; return NULL; }

    for (;;) {
        EB_Error_Code rc = eb_read_binary(book, sizeof(buf), buf, &chunk);
        if (rc != EB_SUCCESS || chunk <= 0) break;
        if (total + (size_t)chunk >= alloc) {
            alloc = (total + (size_t)chunk) * 2 + 1;
            char *tmp = (char *)realloc(result, alloc);
            if (!tmp) { free(result); *out_len = 0; return NULL; }
            result = tmp;
        }
        memcpy(result + total, buf, (size_t)chunk);
        total += (size_t)chunk;
    }
    *out_len = total;
    return result;
}
