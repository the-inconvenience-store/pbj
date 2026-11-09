#!/usr/bin/env -S osascript -l JavaScript

// Verification-only, read-only observer. Call only after an authorized fixture copy.
ObjC.import("AppKit");
ObjC.import("Foundation");

function run(args) {
    if (args.length !== 1) {
        throw new Error("usage: scripts/inspect-pasteboard.js TYPE|--types");
    }
    const items = $.NSPasteboard.generalPasteboard.pasteboardItems;
    const result = [];
    for (let index = 0; index < Number(items.count); index++) {
        const item = items.objectAtIndex(index);
        if (args[0] === "--types") {
            const types = [];
            for (let position = 0; position < Number(item.types.count); position++) {
                types.push(ObjC.unwrap(item.types.objectAtIndex(position)));
            }
            result.push(types);
        } else {
            const data = item.dataForType($(args[0]));
            result.push(data.isNil() ? null : ObjC.unwrap(data.base64EncodedStringWithOptions(0)));
        }
    }
    return JSON.stringify(result);
}
