package triptych

func MakeRingWithReal(N int, realSK []byte) ([]*Point, error) {
	realPK := PubKeyFromSecret(realSK)
	ring := make([]*Point, 0, N)
	for i := 0; i < N-1; i++ {
		_, pk := GenerateKey()
		ring = append(ring, pk)
	}
	ring = append(ring, realPK)

	for i := len(ring) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		ring[i], ring[j] = ring[j], ring[i]
	}
	return ring, nil
}
