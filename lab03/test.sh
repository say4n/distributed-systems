#! /bin/bash

# Usage: ./test.sh

osascript -e '
set numInstances to 5

tell application "iTerm2"
    tell current window
        set configFiles to {"configFile_01.txt", "configFile_23.txt", "configFile_43.txt", "configFile_55.txt", "configFile_02.txt", "configFile_24.txt", "configFile_44.txt", "configFile_56.txt", "configFile_11.txt", "configFile_31.txt", "configFile_45.txt", "configFile_57.txt", "configFile_12.txt", "configFile_32.txt", "configFile_46.txt", "configFile_58.txt", "configFile_13.txt", "configFile_33.txt", "configFile_51.txt", "configFile_59.txt", "configFile_14.txt", "configFile_34.txt", "configFile_52.txt", "configFile_21.txt", "configFile_41.txt", "configFile_53.txt", "configFile_22.txt", "configFile_42.txt", "configFile_54.txt"}

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
