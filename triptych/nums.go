package triptych

import "math/big"

func getNUMS(idx int) *Point {
	seed := make([]byte, 0, 8)
	seed = append(seed, []byte("NUMS")...)
	seed = append(seed, byte((idx>>24)&0xff), byte((idx>>16)&0xff), byte((idx>>8)&0xff), byte(idx&0xff))
	x := hashToField(seed)
	one := big.NewInt(1)
	for {
		rhs := modAdd(modMul(x, modMul(x, x, secpP), secpP), secpB, secpP)
		y, ok := sqrtModP(rhs)
		if ok {
			if y.Bit(0) == 1 {
				y = new(big.Int).Sub(secpP, y)
			}
			return NewPoint(x, y)
		}
		x = modAdd(x, one, secpP)
	}
}

func getMatrixNUMS(rows, cols int) [][]*Point {
	pts := make([][]*Point, rows)
	for i := 0; i < rows; i++ {
		pts[i] = make([]*Point, cols)
		for j := 0; j < cols; j++ {
			pts[i][j] = getNUMS(i*cols + j)
		}
	}
	return pts
}

var JPoint = getNUMS(254)
