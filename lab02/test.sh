#! /bin/bash

osascript -e 'tell application "iTerm2"
    tell current window
        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run echo.go -config config/configFile_6001.txt 2> 6001.log"
            write text "cat 6001.log"
        end tell

        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run echo.go -config config/configFile_6002.txt 2> 6002.log"
            write text "cat 6002.log"
        end tell

        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run echo.go -config config/configFile_6003.txt 2> 6003.log"
            write text "cat 6003.log"
        end tell

        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run echo.go -config config/configFile_6004.txt 2> 6004.log"
            write text "cat 6004.log"
        end tell

        set newTab to (create tab with default profile)
        tell current session of newTab
            write text "go run echo.go -config config/configFile_6005.txt 2> 6005.log"
            write text "cat 6005.log"
        end tell
    end tell
end tell'
