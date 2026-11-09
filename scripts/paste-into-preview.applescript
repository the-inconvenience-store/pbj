#!/usr/bin/osascript

-- Verification-owned helper. The source file is generated, never a user document.
on run argv
    if (count argv) is not 3 then error "usage: paste-into-preview.applescript SOURCE.png SCREENSHOT CAPTURE_HELPER"
    set sourceWindowID to 0
    set pastedWindowID to 0
    tell application "Preview" to set originalWindows to id of every window
    do shell script "/usr/bin/open -a Preview " & quoted form of (item 1 of argv)
    try
        repeat 50 times
            tell application "Preview"
                if (count windows) is not 0 then
                    set candidate to id of front window
                    if candidate is not in originalWindows then set sourceWindowID to candidate
                end if
            end tell
            if sourceWindowID is not 0 then exit repeat
            delay 0.1
        end repeat
        if sourceWindowID is 0 then error "Preview did not open the generated source fixture"
        tell application "Preview" to activate
        tell application "System Events" to tell process "Preview" to keystroke "n" using command down
        repeat 50 times
            tell application "Preview"
                set candidate to id of front window
                if candidate is not sourceWindowID and candidate is not in originalWindows then set pastedWindowID to candidate
            end tell
            if pastedWindowID is not 0 then exit repeat
            delay 0.1
        end repeat
        if pastedWindowID is 0 then error "Preview New from Clipboard did not create a document"
        tell application "Preview"
            if id of front window is not pastedWindowID then error "Generated Preview paste lost focus"
            set pastedTitle to name of front window
        end tell
        do shell script "/usr/bin/osascript -l JavaScript " & quoted form of (item 3 of argv) & " Preview " & quoted form of pastedTitle & " " & quoted form of (item 2 of argv)
        tell application "Preview"
            close window id pastedWindowID saving no
            close window id sourceWindowID saving no
        end tell
        return "Preview New from Clipboard rendered the generated image"
    on error failureMessage number failureNumber
        tell application "Preview"
            try
                if pastedWindowID is not 0 then close window id pastedWindowID saving no
            end try
            try
                if sourceWindowID is not 0 then close window id sourceWindowID saving no
            end try
        end tell
        error failureMessage number failureNumber
    end try
end run
