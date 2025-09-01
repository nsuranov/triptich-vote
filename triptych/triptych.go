package triptych

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

type Signature struct {
	CommA *Point
	CommB *Point
	CommC *Point
	CommD *Point
	X     []*Point     // len m
	Y     []*Point     // len m
	F     [][]*big.Int // m x (n-1) при подписи; верификатор восстанавливает первый столбец
	ZA    *big.Int
	ZC    *big.Int
	Z     *big.Int
	U     *Point // key image
}

// детерминированный transcript hash
func transcriptHash(commA, commB, commC, commD *Point, X, Y []*Point, ring []*Point, message []byte) *big.Int {
	var buf bytes.Buffer
	buf.Write(commA.BytesCompressed())
	buf.Write(commB.BytesCompressed())
	buf.Write(commC.BytesCompressed())
	buf.Write(commD.BytesCompressed())
	for _, p := range X {
		buf.Write(p.BytesCompressed())
	}
	for _, p := range Y {
		buf.Write(p.BytesCompressed())
	}
	for _, p := range ring {
		buf.Write(p.BytesCompressed())
	}
	buf.Write(message)
	h := sha256.Sum256(buf.Bytes())
	return new(big.Int).Mod(new(big.Int).SetBytes(h[:]), secpN)
}

func triptychGetSigma(n, m, l int) [][]*big.Int {
	sigma := make([][]*big.Int, m)
	lDigits := naryDecomp(l, n, m)
	for j := 0; j < m; j++ {
		sigma[j] = make([]*big.Int, n)
		for i := 0; i < n; i++ {
			if delta(lDigits[j], i) == 1 {
				sigma[j][i] = big.NewInt(1)
			} else {
				sigma[j][i] = big.NewInt(0)
			}
		}
	}
	return sigma
}

func triptychGetA(n, m int) (*Point, *big.Int, [][]*big.Int) {
	matrix := grwzsr(n, m)
	r := randScalar()
	C := matrixPedersenCommit(matrix, r)
	return C, r, matrix
}

func triptychGetB(n, m, l int) (*Point, *big.Int, [][]*big.Int) {
	sigma := triptychGetSigma(n, m, l)
	r := randScalar()
	C := matrixPedersenCommit(sigma, r)
	return C, r, sigma
}

func triptychGetC(matrixA, matrixS [][]*big.Int) (*Point, *big.Int, [][]*big.Int) {
	m := len(matrixA)
	n := len(matrixA[0])
	matrixC := make([][]*big.Int, m)
	for j := 0; j < m; j++ {
		matrixC[j] = make([]*big.Int, n)
		for i := 0; i < n; i++ {
			t := scalarSub(big.NewInt(1), scalarMul(big.NewInt(2), matrixS[j][i]))
			matrixC[j][i] = scalarMul(matrixA[j][i], t)
		}
	}
	r := randScalar()
	C := matrixPedersenCommit(matrixC, r)
	return C, r, matrixC
}

func triptychGetD(matrixA [][]*big.Int) (*Point, *big.Int, [][]*big.Int) {
	m := len(matrixA)
	n := len(matrixA[0])
	matrixD := make([][]*big.Int, m)
	for i := 0; i < m; i++ {
		matrixD[i] = make([]*big.Int, n)
		for j := 0; j < n; j++ {
			matrixD[i][j] = scalarSub(big.NewInt(0), scalarMul(matrixA[i][j], matrixA[i][j]))
		}
	}
	r := randScalar()
	C := matrixPedersenCommit(matrixD, r)
	return C, r, matrixD
}

func triptychGetF(matrixS, matrixA [][]*big.Int, x *big.Int) [][]*big.Int {
	m := len(matrixS)
	n := len(matrixS[0])
	fs := make([][]*big.Int, m)
	for j := 0; j < m; j++ {
		fs[j] = make([]*big.Int, n-1)
		for i := 1; i < n; i++ {
			t := scalarAdd(scalarMul(matrixS[j][i], x), matrixA[j][i])
			fs[j][i-1] = t
		}
	}
	return fs
}

func triptychGetX(polys [][]*big.Int, ring []*Point, rhos []*big.Int) []*Point {
	m := len(rhos)
	ptsJ := make([]*Point, m)
	for j := 0; j < m; j++ {
		var sum []*Point
		for k := 0; k < len(ring); k++ {
			if polys[k][j].Sign() == 0 {
				continue
			}
			sum = append(sum, pointScalarMult(polys[k][j], ring[k]))
		}
		sum = append(sum, pointScalarMult(rhos[j], NewPoint(Gx, Gy)))
		ptsJ[j] = pointsSum(sum)
	}
	return ptsJ
}

func triptychGetY(rhos []*big.Int) []*Point {
	out := make([]*Point, len(rhos))
	for i := 0; i < len(rhos); i++ {
		out[i] = pointScalarMult(rhos[i], JPoint)
	}
	return out
}

// RingSignTriptych — подпись.
// seckey: 32 байта; message: произвольные; ring: массив публичных ключей (n^m штук, включая реальный)
func RingSignTriptych(seckey []byte, message []byte, ring []*Point, n, m int) (*Signature, []*Point, error) {
	N := 1
	for i := 0; i < m; i++ {
		N *= n
	}
	if len(ring) != N {
		return nil, nil, ErrRingSize{Need: N, Got: len(ring)}
	}

	realPub := PubKeyFromSecret(seckey)
	found := -1
	for i, p := range ring {
		if bytes.Equal(p.BytesCompressed(), realPub.BytesCompressed()) {
			found = i
			break
		}
	}
	if found == -1 {
		return nil, nil, ErrNoRealKey
	}

	// случайная перестановка кольца
	ringSh := make([]*Point, len(ring))
	copy(ringSh, ring)
	for i := len(ringSh) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		ringSh[i], ringSh[j] = ringSh[j], ringSh[i]
	}

	l := -1
	for i, p := range ringSh {
		if bytes.Equal(p.BytesCompressed(), realPub.BytesCompressed()) {
			l = i
			break
		}
	}

	commA, randA, matrixA := triptychGetA(n, m)
	commB, randB, matrixS := triptychGetB(n, m, l)
	commC, randC, _ := triptychGetC(matrixA, matrixS)
	commD, randD, _ := triptychGetD(matrixA)

	polys := make([][]*big.Int, len(ringSh))
	for i := 0; i < len(ringSh); i++ {
		idigits := naryDecomp(i, n, m)
		poly := []*big.Int{big.NewInt(1)}
		for j := 0; j < m; j++ {
			a := matrixS[j][idigits[j]]
			b := matrixA[j][idigits[j]]
			poly = polyMultLin(poly, a, b)
		}
		for L, R := 0, len(poly)-1; L < R; L, R = L+1, R-1 {
			poly[L], poly[R] = poly[R], poly[L]
		}
		polys[i] = poly
	}

	rhos := make([]*big.Int, m)
	for j := 0; j < m; j++ {
		rhos[j] = randScalar()
	}
	X := triptychGetX(polys, ringSh, rhos)
	Y := triptychGetY(rhos)

	x := transcriptHash(commA, commB, commC, commD, X, Y, ringSh, message)
	f := triptychGetF(matrixS, matrixA, x)

	zA := scalarAdd(randA, scalarMul(x, randB))
	zC := scalarAdd(scalarMul(randC, x), randD)

	xPow := big.NewInt(1)
	sumRho := big.NewInt(0)
	for j := 0; j < m; j++ {
		if j == 0 {
			xPow = big.NewInt(1)
		} else {
			xPow = scalarMul(xPow, x)
		}
		sumRho = scalarAdd(sumRho, scalarMul(xPow, rhos[j]))
	}
	xm := scalarPow(x, m)
	sk := scalarFromBytes32(seckey)
	z := scalarSub(scalarMul(sk, xm), sumRho)
	U := pointScalarMult(sk, JPoint)

	sig := &Signature{
		CommA: commA, CommB: commB, CommC: commC, CommD: commD,
		X: X, Y: Y, F: f, ZA: zA, ZC: zC, Z: z, U: U,
	}
	return sig, ringSh, nil
}

// VerifyTriptych теперь возвращает (ok, uNumCompressed),
// где uNum — это compressed key image U (33 байта), стабильный для одного и того же секретного ключа.
func VerifyTriptych(sig *Signature, message []byte, ring []*Point, n, m int) (bool, []byte) {
	commA, commB, commC, commD := sig.CommA, sig.CommB, sig.CommC, sig.CommD
	X, Y := sig.X, sig.Y
	f := deepCopyMatrix(sig.F)
	zA, zC, z, U := sig.ZA, sig.ZC, sig.Z, sig.U

	x := transcriptHash(commA, commB, commC, commD, X, Y, ring, message)

	for j := 0; j < m; j++ {
		sumRow := big.NewInt(0)
		for i := 0; i < len(f[j]); i++ {
			sumRow = scalarAdd(sumRow, f[j][i])
		}
		first := scalarSub(x, sumRow)
		f[j] = append([]*big.Int{first}, f[j]...)
	}

	lhs1 := pointAdd(commA, pointScalarMult(x, commB))
	rhs1 := matrixPedersenCommit(f, zA)
	if !PointsEqual(lhs1, rhs1) {
		return false, nil
	}

	fxf := make([][]*big.Int, len(f))
	for j := 0; j < len(f); j++ {
		fxf[j] = make([]*big.Int, len(f[0]))
		for i := 0; i < len(f[0]); i++ {
			fxf[j][i] = scalarMul(f[j][i], scalarSub(x, f[j][i]))
		}
	}
	lhs2 := pointAdd(pointScalarMult(x, commC), commD)
	rhs2 := matrixPedersenCommit(fxf, zC)
	if !PointsEqual(lhs2, rhs2) {
		return false, nil
	}

	sum_m_f_terms := NewInfinity()
	sum_u_f_terms := NewInfinity()
	for k := 0; k < len(ring); k++ {
		idigits := naryDecomp(k, n, m)
		prodf := big.NewInt(1)
		for j := 0; j < m; j++ {
			prodf = scalarMul(prodf, f[j][idigits[j]])
		}
		sum_m_f_terms = pointAdd(sum_m_f_terms, pointScalarMult(prodf, ring[k]))
		sum_u_f_terms = pointAdd(sum_u_f_terms, pointScalarMult(prodf, U))
	}

	zG := pointScalarMult(z, NewPoint(Gx, Gy))
	zJ := pointScalarMult(z, JPoint)

	xPow := big.NewInt(1)
	xX := NewInfinity()
	xY := NewInfinity()
	for j := 0; j < m; j++ {
		if j == 0 {
			xPow = big.NewInt(1)
		} else {
			xPow = scalarMul(xPow, x)
		}
		xX = pointAdd(xX, pointScalarMult(xPow, X[j]))
		xY = pointAdd(xY, pointScalarMult(xPow, Y[j]))
	}

	if !PointsEqual(sum_m_f_terms, pointAdd(xX, zG)) {
		return false, nil
	}
	if !PointsEqual(sum_u_f_terms, pointAdd(xY, zJ)) {
		return false, nil
	}

	// успех — возвращаем uNum как compressed U
	return true, U.BytesCompressed()
}

// ошибки
type ErrRingSize struct{ Need, Got int }

func (e ErrRingSize) Error() string { return "ring length must be n^m" }

var ErrNoRealKey = errorsNew("ring must contain the signer pubkey")

// локальная, чтобы не тянуть "errors" во множество файлов
func errorsNew(s string) error { return &stringError{s} }

type stringError struct{ s string }

func (e *stringError) Error() string { return e.s }
