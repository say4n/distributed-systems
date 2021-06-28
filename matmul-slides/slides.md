---
theme: apple-basic
layout: intro
---

# Matrix Multiplication
with Map Reduce

<div class="absolute bottom-10">
    <span class="font-700">
        Sayan Goswami
    </span>
</div>

---

# Matrix Multiplication

Let us consider two matrices $A \in \mathbb{R}^{m \times n}$ and $B \in \mathbb{R}^{n \times o}$.

Their product is a matrix $C \in \mathbb{R}^{m \times o}$.

$$C_{i, j} = \sum_k A_{i, k} B_{k, j}$$

where, $i \in [m]$, $j \in [o]$ and $k \in [n]$.

---

# Matrix Multiplication as Map-Reduce : Stage I

## Map

- Emits items of the form $\langle ijk, A_{ij} \rangle \forall k \in [o]$ for matrix A.
- Emits items of the form $\langle ijk, B_{kj} \rangle \forall i \in [m]$ for matrix B.
	- Note: $B_{kj}$ and not $B_{jk}$.
    - It is flipped as the elements are accessed column wise for matrix B.

## Reduce

- Two values have same key $ijk$, they are from $A_{ik}$ and $B_{kj}$.
- Emits the product of these two with the same key $ijk$.
- In other words, $\langle ijk, A_{ik} \times B_{kj} \rangle$.


---

# Matrix Multiplication as Map-Reduce : Stage II

## Map

- Emits items $\langle ik, A_{ik} \times B_{kj} \rangle$ from input $\langle ijk, A_{ik} \times B_{kj} \rangle$

## Reduce

- Takes values having the same key $ik$ and $sums$ them.
- Emits the result as $\langle ik, sum\rangle$.
- This is the final output.
