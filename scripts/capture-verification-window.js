#!/usr/bin/env -S osascript -l JavaScript

// Capture only a named fixture window, never the whole desktop.
ObjC.import("CoreGraphics");
const host = Application.currentApplication();
host.includeStandardAdditions = true;

function shellQuote(value) {
    return "'" + value.replace(/'/g, "'\\''") + "'";
}

function run(args) {
    if (args.length !== 3) {
        throw new Error("usage: capture-verification-window.js APP FIXTURE_TITLE OUTPUT.png");
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
    host.doShellScript("/usr/sbin/screencapture -x -l " + identifier + " " + shellQuote(args[2]));
    return JSON.stringify({window: identifier});
}
