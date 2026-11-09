#include <errno.h>
#include <iconv.h>
#include <langinfo.h>
#include <locale.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <xlocale.h>
#include "pasteboard.h"

static void pbj_text_error(char **error, const char *operation) {
    int saved_errno = errno;
    if (asprintf(error, "%s: %s", operation, strerror(saved_errno)) < 0) {
        *error = NULL;
    }
}

char *pbj_text_encoding(const char *locale, char **error) {
    locale_t selected = newlocale(LC_CTYPE_MASK, locale, NULL);
    if (selected == NULL) {
        selected = newlocale(LC_CTYPE_MASK, "C", NULL);
    }
    if (selected == NULL) {
        pbj_text_error(error, "create text locale");
        return NULL;
    }
    const char *codeset = nl_langinfo_l(CODESET, selected);
    if (codeset == NULL || codeset[0] == '\0') {
        freelocale(selected);
        *error = strdup("text locale has no character encoding");
        return NULL;
    }
    char *result = strdup(codeset);
    freelocale(selected);
    if (result == NULL) {
        pbj_text_error(error, "copy text encoding");
    }
    return result;
}

int pbj_text_convert(const void *data, size_t length, const char *encoding,
                     void **output, size_t *output_length, char **error) {
    iconv_t converter = iconv_open("UTF-8", encoding);
    if (converter == (iconv_t)-1) {
        pbj_text_error(error, "open text converter");
        return 0;
    }
    size_t capacity = length < 64 ? 64 : length;
    char *buffer = malloc(capacity);
    if (buffer == NULL) {
        pbj_text_error(error, "allocate converted text");
        iconv_close(converter);
        return 0;
    }
    char *input = (char *)data;
    size_t remaining = length;
    size_t written = 0;
    int flushing = 0;
    for (;;) {
        char *destination = buffer + written;
        size_t available = capacity - written;
        size_t result = flushing
            ? iconv(converter, NULL, NULL, &destination, &available)
            : iconv(converter, &input, &remaining, &destination, &available);
        written = capacity - available;
        if (result != (size_t)-1) {
            if (result != 0) {
                *error = strdup("text conversion would substitute characters");
                break;
            }
            if (!flushing) {
                flushing = 1;
                continue;
            }
            iconv_close(converter);
            *output = buffer;
            *output_length = written;
            return 1;
        }
        if (errno != E2BIG) {
            pbj_text_error(error, "invalid text for locale encoding");
            break;
        }
        if (capacity > SIZE_MAX / 2) {
            *error = strdup("converted text exceeds addressable memory");
            break;
        }
        size_t next_capacity = capacity * 2;
        char *grown = realloc(buffer, next_capacity);
        if (grown == NULL) {
            pbj_text_error(error, "grow converted text");
            break;
        }
        buffer = grown;
        capacity = next_capacity;
    }
    free(buffer);
    iconv_close(converter);
    return 0;
}
