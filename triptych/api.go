package triptych

// ВНИМАНИЕ: учебная реализация, не constant-time и не для продакшена.

func GenerateKey() (sk32 []byte, pk *Point) {
	k := randScalar()
	pk = baseScalarMult(k)
	b := k.Bytes()
	if len(b) < 32 {
		pad := make([]byte, 32-len(b))
		b = append(pad, b...)
	}
	return b, pk
}

func PubKeyFromSecret(sk []byte) *Point {
	k := scalarFromBytes32(sk)
	return baseScalarMult(k)
}
