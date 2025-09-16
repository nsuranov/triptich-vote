package triptych

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"math/big"
)

var (
	secpP, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)

	secpN, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

	secpB = big.NewInt(7)

	Gx, _ = new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	Gy, _ = new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
)

type Point struct {
	X, Y *big.Int
	Inf  bool
}

func NewInfinity() *Point { return &Point{Inf: true} }
func NewPoint(x, y *big.Int) *Point {
	return &Point{X: new(big.Int).Set(x), Y: new(big.Int).Set(y), Inf: false}
}

func isOnCurve(P *Point) bool {
	if P == nil || P.Inf {
		return true
	}
	y2 := new(big.Int).Mul(P.Y, P.Y)
	y2.Mod(y2, secpP)
	x2 := new(big.Int).Mul(P.X, P.X)
	x3 := new(big.Int).Mul(x2, P.X)
	x3.Mod(x3, secpP)
	rhs := new(big.Int).Add(x3, secpB)
	rhs.Mod(rhs, secpP)
	return y2.Cmp(rhs) == 0
}

func modInv(a, mod *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, mod)
}
func modAdd(a, b, mod *big.Int) *big.Int {
	z := new(big.Int).Add(a, b)
	z.Mod(z, mod)
	return z
}
func modSub(a, b, mod *big.Int) *big.Int {
	z := new(big.Int).Sub(a, b)
	z.Mod(z, mod)
	if z.Sign() < 0 {
		z.Add(z, mod)
	}
	return z
}
func modMul(a, b, mod *big.Int) *big.Int {
	z := new(big.Int).Mul(a, b)
	z.Mod(z, mod)
	return z
}
func modPow(a, e, mod *big.Int) *big.Int { return new(big.Int).Exp(a, e, mod) }

func sqrtModP(a *big.Int) (*big.Int, bool) {
	ls := new(big.Int).Exp(a, new(big.Int).Rsh(new(big.Int).Sub(secpP, big.NewInt(1)), 1), secpP)
	if ls.Sign() == 0 {
		return big.NewInt(0), true
	}
	if ls.Cmp(big.NewInt(1)) != 0 {
		return nil, false
	}
	exp := new(big.Int).Rsh(new(big.Int).Add(secpP, big.NewInt(1)), 2)
	y := new(big.Int).Exp(a, exp, secpP)
	if new(big.Int).Mod(new(big.Int).Mul(y, y), secpP).Cmp(new(big.Int).Mod(a, secpP)) == 0 {
		return y, true
	}
	return nil, false
}

func pointAdd(P, Q *Point) *Point {
	if P == nil || P.Inf {
		return &Point{X: new(big.Int).Set(Q.X), Y: new(big.Int).Set(Q.Y), Inf: Q.Inf}
	}
	if Q == nil || Q.Inf {
		return &Point{X: new(big.Int).Set(P.X), Y: new(big.Int).Set(P.Y), Inf: P.Inf}
	}
	if P.X.Cmp(Q.X) == 0 {
		sumY := modAdd(P.Y, Q.Y, secpP)
		if sumY.Sign() == 0 {
			return NewInfinity()
		}
		return pointDouble(P)
	}
	num := modSub(Q.Y, P.Y, secpP)
	den := modSub(Q.X, P.X, secpP)
	l := modMul(num, modInv(den, secpP), secpP)
	x3 := modSub(modSub(modMul(l, l, secpP), P.X, secpP), Q.X, secpP)
	y3 := modSub(modMul(l, modSub(P.X, x3, secpP), secpP), P.Y, secpP)
	return NewPoint(x3, y3)
}

func pointDouble(P *Point) *Point {
	if P == nil || P.Inf {
		return NewInfinity()
	}
	if P.Y.Sign() == 0 {
		return NewInfinity()
	}
	threeX2 := modMul(big.NewInt(3), modMul(P.X, P.X, secpP), secpP)
	twoY := modMul(big.NewInt(2), P.Y, secpP)
	l := modMul(threeX2, modInv(twoY, secpP), secpP)
	x3 := modSub(modMul(l, l, secpP), modMul(big.NewInt(2), P.X, secpP), secpP)
	y3 := modSub(modMul(l, modSub(P.X, x3, secpP), secpP), P.Y, secpP)
	return NewPoint(x3, y3)
}

func pointScalarMult(k *big.Int, P *Point) *Point {
	if P == nil || P.Inf {
		return NewInfinity()
	}
	if k.Sign() == 0 {
		return NewInfinity()
	}
	k = new(big.Int).Mod(new(big.Int).Set(k), secpN)
	R := NewInfinity()
	Q := &Point{X: new(big.Int).Set(P.X), Y: new(big.Int).Set(P.Y), Inf: P.Inf}
	for i := k.BitLen() - 1; i >= 0; i-- {
		R = pointDouble(R)
		if k.Bit(i) == 1 {
			R = pointAdd(R, Q)
		}
	}
	return R
}

func baseScalarMult(k *big.Int) *Point { return pointScalarMult(k, NewPoint(Gx, Gy)) }

func pointsSum(points []*Point) *Point {
	acc := NewInfinity()
	for _, p := range points {
		if p == nil {
			continue
		}
		acc = pointAdd(acc, p)
	}
	return acc
}

func (P *Point) BytesCompressed() []byte {
	if P == nil || P.Inf {
		out := make([]byte, 33)
		return out
	}
	prefix := byte(0x02)
	if new(big.Int).Mod(P.Y, big.NewInt(2)).Cmp(big.NewInt(1)) == 0 {
		prefix = 0x03
	}
	x := P.X.Bytes()
	if len(x) < 32 {
		pad := make([]byte, 32-len(x))
		x = append(pad, x...)
	}
	return append([]byte{prefix}, x...)
}

func ParseCompressed(b []byte) (*Point, error) {
	if len(b) != 33 {
		return nil, errors.New("compressed key must be 33 bytes")
	}
	if bytes.Equal(b, make([]byte, 33)) {
		return NewInfinity(), nil
	}
	prefix := b[0]
	if prefix != 0x02 && prefix != 0x03 {
		return nil, errors.New("invalid prefix")
	}
	x := new(big.Int).SetBytes(b[1:])
	rhs := modAdd(modMul(x, modMul(x, x, secpP), secpP), secpB, secpP)
	y, ok := sqrtModP(rhs)
	if !ok {
		return nil, errors.New("not on curve")
	}
	if y.Bit(0) != uint(prefix&1) {
		y = new(big.Int).Sub(secpP, y)
	}
	P := NewPoint(x, y)
	if !isOnCurve(P) {
		return nil, errors.New("point not on curve")
	}
	return P, nil
}

func PointsEqual(a, b *Point) bool {
	if a.Inf && b.Inf {
		return true
	}
	if a.Inf != b.Inf {
		return false
	}
	return a.X.Cmp(b.X) == 0 && a.Y.Cmp(b.Y) == 0
}

func hashToField(seed []byte) *big.Int {
	h := sha256.Sum256(seed)
	x := new(big.Int).SetBytes(h[:])
	x.Mod(x, secpP)
	return x
}
