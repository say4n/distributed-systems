#! /bin/bash

# Usage: ./test.sh

osascript -e '
set numInstances to 5

tell application "iTerm2"
    tell current window
        set configFiles to {"config/configFile_01.txt", "config/configFile_23.txt", "config/configFile_43.txt", "config/configFile_55.txt", "config/configFile_02.txt", "config/configFile_24.txt", "config/configFile_44.txt", "config/configFile_56.txt", "config/configFile_11.txt", "config/configFile_31.txt", "config/configFile_45.txt", "config/configFile_57.txt", "config/configFile_12.txt", "config/configFile_32.txt", "config/configFile_46.txt", "config/configFile_58.txt", "config/configFile_13.txt", "config/configFile_33.txt", "config/configFile_51.txt", "config/configFile_59.txt", "config/configFile_14.txt", "config/configFile_34.txt", "config/configFile_52.txt", "config/configFile_21.txt", "config/configFile_41.txt", "config/configFile_53.txt", "config/configFile_22.txt", "config/configFile_42.txt", "config/configFile_54.txt"}

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
