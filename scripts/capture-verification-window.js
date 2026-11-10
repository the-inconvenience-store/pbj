#!/usr/bin/env -S osascript -l JavaScript

// Capture only a named fixture window, never the whole desktop.
ObjC.import("CoreGraphics");
ObjC.import("AppKit");
const host = Application.currentApplication();
host.includeStandardAdditions = true;

function shellQuote(value) {
    return "'" + value.replace(/'/g, "'\\''") + "'";
}

function run(args) {
    if (args.length !== 3 && args.length !== 7) {
        throw new Error("usage: capture-verification-window.js APP FIXTURE_TITLE OUTPUT.png [X Y WIDTH HEIGHT]");
    }
    const windows = ObjC.deepUnwrap(ObjC.castRefToObject($.CGWindowListCopyWindowInfo(1, 0)));
    const stem = args[1].replace(/\.[^.]+$/, "");
    const match = windows.find(function (window) {
        const title = String(window.kCGWindowName || "");
        return window.kCGWindowOwnerName === args[0] && Number(window.kCGWindowLayer) === 0 &&
            (title.includes(args[1]) || title.includes(stem));
    });
    if (!match) {
        throw new Error("Generated fixture window is not visible or screen-recording access is missing");
    }
    const identifier = Number(match.kCGWindowNumber);
    if (args.length === 3) {
        host.doShellScript("/usr/sbin/screencapture -x -l " + identifier + " " + shellQuote(args[2]));
    } else {
        const bounds = match.kCGWindowBounds;
        const [x, y, width, height] = args.slice(3).map(Number);
        if (![x, y, width, height].every(Number.isFinite) || width <= 0 || height <= 0 ||
            x < bounds.X || y < bounds.Y || x + width > bounds.X + bounds.Width ||
            y + height > bounds.Y + bounds.Height) {
            throw new Error("Body crop is outside the generated fixture window");
        }
        const temporary = args[2] + ".window.png";
        try {
            host.doShellScript("/usr/sbin/screencapture -x -o -l " + identifier + " " + shellQuote(temporary));
            const image = $.NSBitmapImageRep.imageRepWithData($.NSData.dataWithContentsOfFile(temporary));
            const scale = Number(image.pixelsWide) / Number(bounds.Width);
            const rectangle = $.CGRectMake((x - bounds.X) * scale, (y - bounds.Y) * scale,
                                          width * scale, height * scale);
            const crop = $.CGImageCreateWithImageInRect(image.CGImage, rectangle);
            const bitmap = $.NSBitmapImageRep.alloc.initWithCGImage(crop);
            const data = bitmap.representationUsingTypeProperties($.NSBitmapImageFileTypePNG, $.NSDictionary.dictionary);
            if (!data.writeToFileAtomically(args[2], true)) {
                throw new Error("Could not save generated body crop");
            }
        } finally {
            $.NSFileManager.defaultManager.removeItemAtPathError(temporary, null);
        }
    }
    return JSON.stringify({window: identifier});
}
