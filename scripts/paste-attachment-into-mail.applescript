#!/usr/bin/osascript

-- Verification-owned helper. Paste only into an owned, unsent compose window.
on run argv
    if (count argv) is not 4 then error "usage: paste-attachment-into-mail.applescript SUBJECT FILENAME SCREENSHOT CAPTURE_HELPER"
    set fixtureSubject to item 1 of argv
    set expectedName to item 2 of argv
    tell application "Mail"
        set fixture to make new outgoing message with properties {subject:fixtureSubject, content:"Generated pbjcopy attachment verification.\n", visible:true, message signature:missing value}
        activate
    end tell
    try
        set bodyElement to missing value
        repeat 50 times
            tell application "System Events"
                tell process "Mail"
                    set matchingWindows to windows whose name contains fixtureSubject
                    if (count matchingWindows) is not 0 then
                        set targetWindow to item 1 of matchingWindows
                        set bodyElement to my findBody(targetWindow)
                    end if
                end tell
            end tell
            if bodyElement is not missing value then exit repeat
            delay 0.1
        end repeat
        if bodyElement is missing value then error "Generated Mail compose body is unavailable"
        tell application "System Events"
            tell process "Mail"
                set frontmost to true
                repeat 50 times
                    if frontmost then exit repeat
                    delay 0.1
                end repeat
                if not frontmost then error "Mail could not become the foreground application"
                if name of front window does not contain fixtureSubject then error "Generated Mail compose window is not frontmost"
                set value of attribute "AXFocused" of bodyElement to true
                key code 125 using command down
                keystroke "v" using command down
            end tell
        end tell
        set attachmentRole to ""
        repeat 100 times
            set attachmentRole to my findAttachment(bodyElement, expectedName)
            if attachmentRole is not "" then exit repeat
            delay 0.1
        end repeat
        tell application "System Events"
            set {bodyX, bodyY} to position of bodyElement
            set {bodyWidth, bodyHeight} to size of bodyElement
        end tell
        -- Body-only capture excludes account identities and all unrelated windows.
        do shell script "/usr/bin/osascript -l JavaScript " & quoted form of (item 4 of argv) & " Mail " & quoted form of fixtureSubject & " " & quoted form of (item 3 of argv) & " " & (bodyX as integer as text) & " " & (bodyY as integer as text) & " " & (bodyWidth as integer as text) & " " & (bodyHeight as integer as text)
        if attachmentRole is "" then error "Mail did not expose the named pasted attachment"
        tell application "Mail" to close fixture saving no
        return "Mail displayed the named file attachment (" & attachmentRole & "); generated unsent message discarded"
    on error failureMessage number failureNumber
        try
            tell application "Mail" to close fixture saving no
        end try
        error failureMessage number failureNumber
    end try
end run

on findBody(parentElement)
    tell application "System Events"
        if role of parentElement is "AXWebArea" then return parentElement
        repeat with elementIndex from 1 to count UI elements of parentElement
            set childElement to UI element elementIndex of parentElement
            set resultElement to my findBody(childElement)
            if resultElement is not missing value then return resultElement
        end repeat
    end tell
    return missing value
end findBody

on findAttachment(parentElement, expectedName)
    tell application "System Events"
        set elementRole to role of parentElement
        if elementRole is "AXButton" then
            if (description of parentElement as text) contains expectedName then return elementRole
        end if
        repeat with elementIndex from 1 to count UI elements of parentElement
            set childElement to UI element elementIndex of parentElement
            set resultRole to my findAttachment(childElement, expectedName)
            if resultRole is not "" then return resultRole
        end repeat
    end tell
    return ""
end findAttachment
