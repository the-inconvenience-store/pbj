#import <AppKit/AppKit.h>
#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "pasteboard.h"

static void pbj_error(char **error, NSString *message) {
    if (error != NULL) {
        *error = strdup(message.UTF8String ?: "unknown native failure");
    }
}

static void pbj_exception(char **error, NSException *exception) {
    pbj_error(error, [NSString stringWithFormat:@"%@: %@", exception.name,
                     exception.reason ?: @"no reason supplied"]);
}

void *pbj_board_open(const void *name, size_t length, int unique,
                     char **actual_name, char **error) {
    @autoreleasepool {
        @try {
            NSPasteboard *board;
            if (unique) {
                board = [NSPasteboard pasteboardWithUniqueName];
            } else {
                NSString *boardName = [[NSString alloc] initWithBytes:name length:length
                                                           encoding:NSUTF8StringEncoding];
                if (boardName == nil || boardName.length == 0) {
                    pbj_error(error, @"invalid pasteboard name");
                    return NULL;
                }
                if ([boardName isEqualToString:@"general"]) {
                    boardName = NSPasteboardNameGeneral;
                } else if ([boardName isEqualToString:@"find"]) {
                    boardName = NSPasteboardNameFind;
                } else if ([boardName isEqualToString:@"ruler"]) {
                    boardName = NSPasteboardNameRuler;
                } else if ([boardName isEqualToString:@"font"]) {
                    boardName = NSPasteboardNameFont;
                }
                board = [NSPasteboard pasteboardWithName:boardName];
            }
            if (board == nil) {
                pbj_error(error, @"AppKit did not return a pasteboard");
                return NULL;
            }
            *actual_name = strdup(board.name.UTF8String);
            if (*actual_name == NULL) {
                if (unique) {
                    [board releaseGlobally];
                }
                pbj_error(error, @"could not allocate pasteboard name");
                return NULL;
            }
            return (void *)CFBridgingRetain(board);
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            return NULL;
        }
    }
}

void *pbj_item_create(char **error) {
    @autoreleasepool {
        @try {
            NSPasteboardItem *item = [[NSPasteboardItem alloc] init];
            if (item == nil) {
                pbj_error(error, @"could not allocate pasteboard item");
                return NULL;
            }
            return (void *)CFBridgingRetain(item);
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            return NULL;
        }
    }
}

int pbj_item_set(void *item, const void *type, size_t type_length,
                 const void *data, size_t data_length, char **error) {
    @autoreleasepool {
        @try {
            NSString *identifier = [[NSString alloc] initWithBytes:type length:type_length
                                                        encoding:NSUTF8StringEncoding];
            if (identifier == nil || identifier.length == 0 ||
                memchr(type, 0, type_length) != NULL) {
                pbj_error(error, @"invalid pasteboard representation type");
                return 0;
            }
            // Copy directly from the synchronously borrowed Go buffer. AppKit
            // retains eager NSData; no Go pointer escapes this call.
            NSData *bytes = [[NSData alloc] initWithBytes:data length:data_length];
            if (bytes == nil || ![(__bridge NSPasteboardItem *)item setData:bytes forType:identifier]) {
                pbj_error(error, [NSString stringWithFormat:@"AppKit rejected representation %@", identifier]);
                return 0;
            }
            return 1;
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            return 0;
        }
    }
}

static int pbj_publish(NSPasteboard *board, NSArray<NSPasteboardItem *> *items, char **error) {
    [board clearContents];
    if (![board writeObjects:items]) {
        pbj_error(error, @"AppKit writeObjects returned false after clearing the pasteboard");
        return 0;
    }
    // A fresh reader reference acknowledges the global resource, not AppKit's
    // writer-side cache. Without this round trip a short-lived general-board
    // writer can exit while subsequent file items are still being handed off.
    PasteboardRef reader = NULL;
    OSStatus status = PasteboardCreate((__bridge CFStringRef)board.name, &reader);
    if (status != noErr || reader == NULL) {
        pbj_error(error, [NSString stringWithFormat:@"open publication acknowledgment failed (%d)", (int)status]);
        return 0;
    }
    PasteboardSynchronize(reader);
    ItemCount count = 0;
    status = PasteboardGetItemCount(reader, &count);
    CFRelease(reader);
    if (status != noErr) {
        pbj_error(error, [NSString stringWithFormat:@"acknowledge publication failed (%d)", (int)status]);
        return 0;
    }
    // Another application may already have replaced the contents. Do not assert
    // a count or claim ownership: the successful write follows last-writer wins.
    return 1;
}

int pbj_board_write(void *board, void *item, char **error) {
    @autoreleasepool {
        @try {
            return pbj_publish((__bridge NSPasteboard *)board,
                               @[(__bridge NSPasteboardItem *)item], error);
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            return 0;
        }
    }
}

void *pbj_files_create(char **error) {
    @autoreleasepool {
        @try {
            NSMutableArray<NSPasteboardItem *> *items = [[NSMutableArray alloc] init];
            if (items == nil) {
                pbj_error(error, @"could not allocate file-reference items");
                return NULL;
            }
            return (void *)CFBridgingRetain(items);
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            return NULL;
        }
    }
}

int pbj_files_add(void *files, const void *path, size_t length,
                  int directory, char **error) {
    @autoreleasepool {
        @try {
            NSString *filePath = [[NSString alloc] initWithBytes:path length:length
                                                      encoding:NSUTF8StringEncoding];
            if (filePath == nil || !filePath.isAbsolutePath || memchr(path, 0, length) != NULL) {
                pbj_error(error, @"invalid absolute file-reference path");
                return 0;
            }
            // Construct the URL from the supplied path, never from a resolved
            // symlink target. This does not open or read the referenced file.
            NSURL *url = [NSURL fileURLWithPath:filePath isDirectory:directory != 0];
            NSPasteboardItem *item = [[NSPasteboardItem alloc] init];
            if (url == nil || item == nil ||
                ![item setString:url.absoluteString forType:NSPasteboardTypeFileURL]) {
                pbj_error(error, @"could not prepare file-reference representation");
                return 0;
            }
            [(__bridge NSMutableArray<NSPasteboardItem *> *)files addObject:item];
            return 1;
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            return 0;
        }
    }
}

int pbj_files_write(void *board, void *files, char **error) {
    @autoreleasepool {
        @try {
            NSArray<NSPasteboardItem *> *items = (__bridge NSArray<NSPasteboardItem *> *)files;
            if (items.count == 0) {
                pbj_error(error, @"file references must not be empty");
                return 0;
            }
            return pbj_publish((__bridge NSPasteboard *)board, items, error);
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            return 0;
        }
    }
}

pbj_snapshot *pbj_board_read(void *board, char **error) {
    @autoreleasepool {
        pbj_snapshot *snapshot = calloc(1, sizeof(*snapshot));
        if (snapshot == NULL) {
            pbj_error(error, @"could not allocate pasteboard snapshot");
            return NULL;
        }
        @try {
            NSArray<NSPasteboardItem *> *items = [(__bridge NSPasteboard *)board pasteboardItems];
            NSMutableArray<NSData *> *owner = [[NSMutableArray alloc] init];
            snapshot->owner = (void *)CFBridgingRetain(owner);
            snapshot->count = items.count;
            if (snapshot->count != 0) {
                snapshot->items = calloc(snapshot->count, sizeof(*snapshot->items));
                if (snapshot->items == NULL) {
                    pbj_error(error, @"could not allocate pasteboard snapshot items");
                    pbj_snapshot_free(snapshot);
                    return NULL;
                }
            }
            for (NSUInteger i = 0; i < items.count; i++) {
                NSPasteboardItem *item = items[i];
                NSArray<NSPasteboardType> *types = item.types;
                pbj_read_item *readItem = &snapshot->items[i];
                readItem->count = types.count;
                if (readItem->count != 0) {
                    readItem->representations = calloc(readItem->count, sizeof(*readItem->representations));
                    if (readItem->representations == NULL) {
                        pbj_error(error, @"could not allocate snapshot representations");
                        pbj_snapshot_free(snapshot);
                        return NULL;
                    }
                }
                for (NSUInteger j = 0; j < types.count; j++) {
                    NSString *type = types[j];
                    NSData *typeBytes = [type dataUsingEncoding:NSUTF8StringEncoding];
                    NSData *data = [item dataForType:type];
                    if (typeBytes == nil || data == nil) {
                        pbj_error(error, [NSString stringWithFormat:@"could not read representation %@", type]);
                        pbj_snapshot_free(snapshot);
                        return NULL;
                    }
                    [owner addObject:typeBytes];
                    [owner addObject:data];
                    readItem->representations[j] = (pbj_read_representation){
                        .type = typeBytes.bytes,
                        .type_length = typeBytes.length,
                        .data = data.bytes,
                        .data_length = data.length,
                    };
                }
            }
            return snapshot;
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            pbj_snapshot_free(snapshot);
            return NULL;
        }
    }
}

void pbj_snapshot_free(pbj_snapshot *snapshot) {
    if (snapshot == NULL) {
        return;
    }
    @autoreleasepool {
        if (snapshot->items != NULL) {
            for (size_t i = 0; i < snapshot->count; i++) {
                free(snapshot->items[i].representations);
            }
        }
        free(snapshot->items);
        (void)CFBridgingRelease(snapshot->owner);
        free(snapshot);
    }
}

int pbj_board_close(void *board, int unique, char **error) {
    @autoreleasepool {
        NSPasteboard *pasteboard = CFBridgingRelease(board);
        @try {
            if (unique) {
                [pasteboard releaseGlobally];
            }
            return 1;
        } @catch (NSException *exception) {
            pbj_exception(error, exception);
            return 0;
        }
    }
}

void pbj_release(void *object) {
    @autoreleasepool {
        (void)CFBridgingRelease(object);
    }
}
