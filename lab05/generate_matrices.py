import argparse
import random

def main(dim_a, dim_b):
    i, j = dim_a
    j, k = dim_b
    print("Writing matrix A to a.mat")
    with open("a.mat", "wt") as matrix_a_file:
        for row_a in range(dim_a[0]):
            for col_a in range(dim_a[1]):
                # matrix_a_file.write(f"{k},{row_a},{col_a},{random.random() * 100:.2f},1\n")
                if row_a == col_a:
                    matrix_a_file.write(f"{k},{row_a},{col_a},1,1\n")
                else:
                    matrix_a_file.write(f"{k},{row_a},{col_a},0,1\n")

    print("Writing matrix B to b.mat")
    with open("b.mat", "wt") as matrix_b_file:
        for row_b in range(dim_b[0]):
            for col_b in range(dim_b[1]):
                # matrix_b_file.write(f"{i},{row_b},{col_b},{random.random() * 100:.2f},0\n")
                if row_b == col_b:
                    matrix_b_file.write(f"{k},{row_b},{col_b},1,0\n")
                else:
                    matrix_b_file.write(f"{k},{row_b},{col_b},0,0\n")

    print("Done.")



if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument("--ra", default=0, type=int, required=True)
    parser.add_argument("--ca", default=0, type=int, required=True)
    parser.add_argument("--rb", default=0, type=int, required=True)
    parser.add_argument("--cb", default=0, type=int, required=True)

    args = parser.parse_args()

    dim_a =(args.ra, args.ca)
    dim_b =(args.rb, args.cb)

    assert args.ca == args.rb, "m x n matrix can only be multiplied with a n x o matrix."

    main(dim_a, dim_b)