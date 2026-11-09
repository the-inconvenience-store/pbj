#!/usr/bin/osascript

-- Verification-owned helper: use Finder's Paste menu in a generated destination.
on run argv
    if (count argv) is not 3 then error "usage: paste-into-finder.applescript DESTINATION FIRST_FILENAME SECOND_FILENAME"
    set destinationPath to item 1 of argv
    tell application "Finder"
        set targetWindow to make new Finder window
        set target of targetWindow to (POSIX file destinationPath as alias)
        set current view of targetWindow to list view
        activate
    end tell
    try
        tell application "Finder"
            if id of front Finder window is not id of targetWindow then error "Generated Finder destination is not frontmost"
        end tell
        tell application "System Events"
            tell process "Finder"
                set frontmost to true
                repeat 50 times
                    if frontmost then exit repeat
                    delay 0.1
                end repeat
                set pasteEntry to missing value
                repeat with entry in menu items of menu 1 of menu bar item "Edit" of menu bar 1
                    if (name of entry starts with "Paste") and (name of entry does not contain "Exactly") and enabled of entry then
                        set pasteEntry to entry
                        exit repeat
                    end if
                end repeat
                if pasteEntry is missing value then error "Finder Paste is not available for the generated destination"
                click pasteEntry
            end tell
        end tell
        set copiesReady to false
        repeat 100 times
            tell application "Finder"
                set destinationFolder to target of targetWindow
                set firstReady to exists file (item 2 of argv) of destinationFolder
                set secondReady to exists file (item 3 of argv) of destinationFolder
            end tell
            if firstReady and secondReady then
                set copiesReady to true
                exit repeat
            end if
            delay 0.1
        end repeat
        if copiesReady is false then error "Finder did not paste the generated file references"
        tell application "Finder" to close targetWindow
        return "Finder pasted both generated file references into the destination"
    on error failureMessage number failureNumber
        try
            tell application "Finder" to close targetWindow
        end try
        error failureMessage number failureNumber
    end try
end run
