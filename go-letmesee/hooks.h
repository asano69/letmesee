#ifndef HOOKS_H
#define HOOKS_H

#include <eb/eb.h>
#include <eb/error.h>
#include <eb/text.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

/*
 * Context passed to every hook callback via the container parameter.
 * Allocated on the C heap so the GC never moves it.
 */
typedef struct EBHookContext {
    int         book_index;
    const char *index_url;    /* C string, owned by the context */
    int         force_inline; /* 0 = link, 1 = inline <img> for color graphics */
    int         fontsize;     /* logical font size in pixels (16/24/30/48) */
    int         fontsize_n;   /* narrow glyph pixel width */
    int         fontsize_w;   /* wide glyph pixel width  */
    int         decoration;   /* tracks current DECORATION code between begin/end */
    char        dict_params[1024]; /* e.g. ";dict=0;dict=1" for reference links */
} EBHookContext;

/* Register all hook functions onto hookset. */
void register_all_hooks(EB_Hookset *hookset);

/*
 * Read the full text from the current position into a malloc'd buffer.
 * The caller must free() the returned pointer.
 * Returns NULL on error; *out_len is set to the number of bytes written.
 */
char *read_text_full(EB_Book *book, EB_Appendix *appendix,
                     EB_Hookset *hookset, void *ctx, size_t *out_len);

/*
 * Read a single heading from the current position.
 * Same ownership rules as read_text_full.
 */
char *read_heading_once(EB_Book *book, EB_Appendix *appendix,
                        EB_Hookset *hookset, void *ctx, size_t *out_len);

#endif /* HOOKS_H */
