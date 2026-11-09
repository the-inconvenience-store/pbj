#ifndef PBJ_PASTEBOARD_H
#define PBJ_PASTEBOARD_H

#include <stddef.h>

typedef struct {
    const void *type;
    size_t type_length;
} pbj_read_representation;

typedef struct {
    pbj_read_representation *representations;
    size_t count;
} pbj_read_item;

typedef struct {
    pbj_read_item *items;
    size_t count;
    void *owner;
    void *board;
    void *source;
    long change_count;
} pbj_snapshot;

/* Handles are retained Objective-C objects, released exactly once by their owner.
 * Input buffers are borrowed only for the duration of a call. Errors and names
 * are malloc-owned strings; the caller frees them. */
void *pbj_board_open(const void *name, size_t length, int unique,
                     char **actual_name, char **error);
void *pbj_item_create(char **error);
int pbj_item_set(void *item, const void *type, size_t type_length,
                 const void *data, size_t data_length, char **error);
int pbj_board_write(void *board, void *item, char **error);
void *pbj_files_create(char **error);
int pbj_files_add(void *files, const void *path, size_t length,
                  int directory, char **error);
int pbj_files_write(void *board, void *files, char **error);
pbj_snapshot *pbj_board_capture(void *board, char **error);
void *pbj_snapshot_data(pbj_snapshot *snapshot, size_t item, size_t representation,
                        const void **data, size_t *length, int *unavailable, char **error);
void pbj_snapshot_free(pbj_snapshot *snapshot);
int pbj_board_close(void *board, int unique, char **error);
void pbj_release(void *object);

/* Locale and conversion results are malloc-owned. Locale lookup uses only
 * thread-independent locale objects, never process-global setlocale. */
char *pbj_text_encoding(const char *locale, char **error);
int pbj_text_convert(const void *data, size_t length, const char *encoding,
                     void **output, size_t *output_length, char **error);
int pbj_text_encode(const void *data, size_t length, const char *encoding,
                    void **output, size_t *output_length, char **error);

#endif
