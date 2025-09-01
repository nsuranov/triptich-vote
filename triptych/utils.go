package triptych

import (
	"crypto/rand"
	"math/big"
)

func delta(a, b int) int {
	if a == b {
		return 1
	}
	return 0
}

// разложение x в базе n длиной m (младшие разряды первыми)
func naryDecomp(x, n, m int) []int {
	if n == 0 {
		out := make([]int, m)
		return out
	}
	digits := []int{}
	v := x
	for v > 0 {
		digits = append(digits, v%n)
		v /= n
	}
	for len(digits) < m {
		digits = append(digits, 0)
	}
	return digits
}

// умножение полинома на линейный (a x + b), коэффициенты в порядке убывания степеней
func polyMultLin(coeffs []*big.Int, a, b *big.Int) []*big.Int {
	L := len(coeffs)
	out := make([]*big.Int, L+1)
	for i := range out {
		out[i] = big.NewInt(0)
	}
	out[0] = scalarMul(a, coeffs[0])
	for i := 1; i < len(out)-1; i++ {
		t1 := scalarMul(b, coeffs[i-1])
		t2 := scalarMul(a, coeffs[i])
		out[i] = scalarAdd(t1, t2)
	}
	out[len(out)-1] = scalarMul(b, coeffs[L-1])
	return out
}

func randomInt(n int) int {
	if n <= 0 {
		return 0
	}
	var b [8]byte
	_, _ = rand.Read(b[:])
	x := new(big.Int).SetBytes(b[:])
	x.Mod(x, big.NewInt(int64(n)))
	return int(x.Int64())
}

func deepCopyMatrix(mtx [][]*big.Int) [][]*big.Int {
	out := make([][]*big.Int, len(mtx))
	for i := range mtx {
		out[i] = make([]*big.Int, len(mtx[i]))
		for j := range mtx[i] {
			out[i][j] = new(big.Int).Set(mtx[i][j])
		}
	}
	return out
}
