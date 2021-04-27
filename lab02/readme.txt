Submission for Lab 01
======================================
Author  :   Sayan Goswami
Email   :   email@sayan.page (sayan.goswami01@estudiant.upf.edu)


--------------------------------------
Usage instructions for clientserver.go
--------------------------------------

The clientserver.go file expects a config file as an argument. It has to passed
with the `-config path_to_file` flag.

An example usage that would run the program with a config file named
configFile_6001.txt in a directory named config would be:

go run clientserver.go -config config/configFile_6001.txt

Alternatively, an executable can be built with the following command

go build clientserver.go

This binary can then be used by passing the same `-config path_to_file` flag.

To run the solution for the lab, 5 such processes need to be launched in
different terminal windows/tabs with their respective config files.

Once all the 5 instances have been instantiated, messages are sent between them
by typing text in the terminal window, the program checks for input from the
stdin in a non-blocking fashion and forwards the input text to its peer nodes
that were provided in the config files during instantiation.
