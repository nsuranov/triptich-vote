package triptych

import (
	"crypto/rand"
	"math/big"
)

func randScalar() *big.Int {
	for {
		var buf [32]byte
		_, _ = rand.Read(buf[:])
		k := new(big.Int).SetBytes(buf[:])
		k.Mod(k, secpN)
		if k.Sign() != 0 {
			return k
		}
	}
}

func scalarFromBytes32(b []byte) *big.Int {
	k := new(big.Int).SetBytes(b)
	k.Mod(k, secpN)
	return k
}

func scalarAdd(a, b *big.Int) *big.Int { return modAdd(a, b, secpN) }
func scalarSub(a, b *big.Int) *big.Int { return modSub(a, b, secpN) }
func scalarMul(a, b *big.Int) *big.Int { return modMul(a, b, secpN) }

func scalarPow(a *big.Int, e int) *big.Int {
	if e == 0 {
		return big.NewInt(1)
	}
	exp := new(big.Int).SetInt64(int64(e))
	return modPow(a, exp, secpN)
}
