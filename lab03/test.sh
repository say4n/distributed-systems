#! /bin/bash

# Usage: ./test.sh

osascript -e '
set numInstances to 5

tell application "iTerm2"
    tell current window
        set configFiles to {"config/configFile_6001.txt", "config/configFile_6003.txt", "config/configFile_6005.txt", "config/configFile_6002.txt", "config/configFile_6004.txt"}

        repeat with cfgFile in configFiles
            set newTab to (create tab with default profile)
            tell current session of newTab
                set runcmd to "go run election.go -config " & cfgfile
                write text runcmd
            end tell
        end repeat
    end tell
end tell
'
