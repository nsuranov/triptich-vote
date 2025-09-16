package triptych

import "math/big"

func grwzsr(n, m int) [][]*big.Int {
	mat := make([][]*big.Int, m)
	for j := 0; j < m; j++ {
		mat[j] = make([]*big.Int, n)
		sum := big.NewInt(0)
		for i := 1; i < n; i++ {
			q := randScalar()
			mat[j][i] = q
			sum = scalarAdd(sum, q)
		}
		mat[j][0] = scalarSub(big.NewInt(0), sum)
	}
	return mat
}

func matrixPedersenCommit(matrix [][]*big.Int, randomness *big.Int) *Point {
	rows := len(matrix)
	cols := len(matrix[0])
	H := getMatrixNUMS(rows, cols)
	var pts []*Point
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if matrix[i][j].Sign() == 0 {
				continue
			}
			pts = append(pts, pointScalarMult(matrix[i][j], H[i][j]))
		}
	}
	pts = append(pts, pointScalarMult(randomness, NewPoint(Gx, Gy)))
	return pointsSum(pts)
}
