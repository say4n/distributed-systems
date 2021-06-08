#! /bin/bash

python3 generate_matrices.py --ra 4 --ca 4 --rb 4 --cb 4
make build
make clean
make execute
