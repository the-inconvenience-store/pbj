#!/usr/bin/osascript

-- Verification-owned helper: opens only the supplied generated fixture.
on run argv
    if (count argv) is not 5 then error "usage: paste-into-textedit.applescript FIXTURE EXPECTED plain|rich SCREENSHOT CAPTURE_HELPER"
    set fixturePath to item 1 of argv
    set expectedText to item 2 of argv
    set richMode to (item 3 of argv is "rich")
    set fixtureName to name of (info for (POSIX file fixturePath))
    -- Launch Services grants the sandboxed editor access to the fixture URL.
    do shell script "/usr/bin/open -a TextEdit " & quoted form of fixturePath
    set targetDocument to missing value
    repeat 50 times
        tell application "TextEdit"
            set matchingDocuments to documents whose name is fixtureName
            if (count matchingDocuments) is not 0 then set targetDocument to item 1 of matchingDocuments
        end tell
        if targetDocument is not missing value then exit repeat
        delay 0.1
    end repeat
    if targetDocument is missing value then error "TextEdit did not open the generated fixture"
    tell application "TextEdit" to activate
    try
        tell application "TextEdit"
            if name of front document is not fixtureName then error "Generated TextEdit fixture is not frontmost"
        end tell
        tell application "System Events"
            tell process "TextEdit"
                set frontmost to true
                keystroke "a" using command down
                keystroke "v" using command down
            end tell
        end tell
        set pastedText to ""
        repeat 50 times
            tell application "TextEdit" to set pastedText to text of targetDocument
            if pastedText contains expectedText then exit repeat
            delay 0.1
        end repeat
        if pastedText does not contain expectedText then error "TextEdit did not paste the generated fixture"
        set pastedFont to "plain"
        tell application "TextEdit"
            if richMode then set pastedFont to font of character 1 of text of targetDocument
            save targetDocument
            if name of front document is not fixtureName then error "Generated TextEdit fixture lost focus"
        end tell
        do shell script "/usr/bin/osascript -l JavaScript " & quoted form of (item 5 of argv) & " TextEdit " & quoted form of fixtureName & " " & quoted form of (item 4 of argv)
        tell application "TextEdit" to close targetDocument saving no
        return pastedText & (ASCII character 31) & pastedFont
    on error failureMessage number failureNumber
        try
            tell application "TextEdit" to close targetDocument saving no
        end try
        error failureMessage number failureNumber
    end try
end run
