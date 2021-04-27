#! /bin/bash

osascript -e 'tell application "iTerm2"
    tell current window
        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run clientserver.go -config config/configFile_6001.txt"
        end tell

        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run clientserver.go -config config/configFile_6002.txt"
        end tell

        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run clientserver.go -config config/configFile_6003.txt"
        end tell

        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run clientserver.go -config config/configFile_6004.txt"
        end tell

        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run clientserver.go -config config/configFile_6005.txt"
        end tell
    end tell
end tell'
