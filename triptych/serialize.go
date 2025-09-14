package triptych

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
)

func Serialize(sig *Signature) (raw []byte, keyImage []byte) {
	var buf bytes.Buffer
	for _, p := range []*Point{sig.CommA, sig.CommB, sig.CommC, sig.CommD} {
		buf.Write(p.BytesCompressed())
	}
	for _, p := range sig.X {
		buf.Write(p.BytesCompressed())
	}
	for _, p := range sig.Y {
		buf.Write(p.BytesCompressed())
	}
	for j := 0; j < len(sig.F); j++ {
		for i := 0; i < len(sig.F[0]); i++ {
			b := sig.F[j][i].Bytes()
			if len(b) < 32 {
				pad := make([]byte, 32-len(b))
				b = append(pad, b...)
			}
			buf.Write(b)
		}
	}
	for _, z := range []*big.Int{sig.ZA, sig.ZC, sig.Z} {
		b := z.Bytes()
		if len(b) < 32 {
			pad := make([]byte, 32-len(b))
			b = append(pad, b...)
		}
		buf.Write(b)
	}
	//_, pk := GenerateKey()
	//fakeU := NewPoint(big.NewInt(1), big.NewInt(2))
	return buf.Bytes(), sig.U.BytesCompressed()
}

// Deserialize — требуется знать m и n (как и при верификации)
// ВАЖНО: keyImg — это сжатая точка U (33 байта), передаваемая отдельно (как у тебя: первые 33 байта перед raw).
func Deserialize(raw []byte, m, n int, keyImg []byte) (*Signature, error) {
	need :=
		4*33 + // A,B,C,D
			m*33 + // X
			m*33 + // Y
			m*(n-1)*32 + // F (без первого столбца)
			3*32 // ZA,ZC,Z
	if len(raw) != need {
		return nil, errors.New("invalid raw length for given m,n")
	}
	if len(keyImg) != 33 {
		return nil, errors.New("key image must be 33 bytes")
	}

	off := 0
	read := func(n int) []byte {
		b := raw[off : off+n]
		off += n
		return b
	}

	p := func(b []byte) *Point {
		P, _ := ParseCompressed(b)
		return P
	}

	sig := &Signature{}
	sig.CommA = p(read(33))
	sig.CommB = p(read(33))
	sig.CommC = p(read(33))
	sig.CommD = p(read(33))

	sig.X = make([]*Point, m)
	for i := 0; i < m; i++ {
		sig.X[i] = p(read(33))
	}
	sig.Y = make([]*Point, m)
	for i := 0; i < m; i++ {
		sig.Y[i] = p(read(33))
	}
	sig.F = make([][]*big.Int, m)
	for j := 0; j < m; j++ {
		sig.F[j] = make([]*big.Int, n-1)
		for i := 0; i < n-1; i++ {
			b := read(32)
			sig.F[j][i] = new(big.Int).Mod(new(big.Int).SetBytes(b), secpN)
		}
	}
	zA := new(big.Int).SetBytes(read(32))
	zC := new(big.Int).SetBytes(read(32))
	z := new(big.Int).SetBytes(read(32))
	sig.ZA = new(big.Int).Mod(zA, secpN)
	sig.ZC = new(big.Int).Mod(zC, secpN)
	sig.Z = new(big.Int).Mod(z, secpN)

	U, err := ParseCompressed(keyImg)
	if err != nil {
		return nil, err
	}
	sig.U = U
	return sig, nil
}

// helpers for CLI
func HexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}
