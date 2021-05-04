#! /bin/bash

# Usage: ./test.sh

osascript -e '
set numInstances to 5

tell application "iTerm2"
    tell current window
        set configFiles to {}

        repeat with n from 1 to numInstances
            set cfgfile to "config/configFile_600" & n & ".txt"
            set configFiles to configFiles & {cfgfile}
        end repeat


        repeat with cfgFile in configFiles
            set newTab to (create tab with default profile)
            tell current session of newTab
                set runcmd to "go run echo.go -config " & cfgfile
                write text runcmd
            end tell
        end repeat
    end tell
end tell
'
