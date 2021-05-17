Submission for Lab 01
======================================
Author  :   Sayan Goswami
Email   :   email@sayan.page (sayan.goswami01@estudiant.upf.edu)


--------------------------------------
Usage instructions for echo.go
--------------------------------------

The election.go file expects a config file as an argument. It has to passed
with the `-config path_to_file` flag.

An example usage that would run the program with a config file named
configFile_6001.txt in a directory named config would be:

go run election.go -config config/configFile_6001.txt

Alternatively, an executable can be built with the following command

go build election.go

This binary can then be used by passing the same `-config path_to_file` flag.

To run the solution for the lab, 5 such processes need to be launched in
different terminal windows/tabs with their respective config files.

Once all the 5 instances have been instantiated, the election algorithm as
described in the problem specification is executed and and a leader is elected,
this leader is printed to stdout before termination.
