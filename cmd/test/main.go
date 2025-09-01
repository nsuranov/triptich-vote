package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"coursach/triptych/triptych"
)

func main() {
	if len(os.Args) >= 4 {
		base, _ := strconv.Atoi(os.Args[1])
		exp, _ := strconv.Atoi(os.Args[2])
		msg := []byte(os.Args[3])

		N := 1
		for i := 0; i < exp; i++ {
			N *= base
		}
		fmt.Printf("Triptych demo: ring size = %d^%d = %d\n", base, exp, N)

		// secret and ring
		seckey, _ := triptych.GenerateKey()
		ring, err := triptych.MakeRingWithReal(N, seckey)
		if err != nil {
			panic(err)
		}

		t0 := time.Now()
		sig, ringUsed, err := triptych.RingSignTriptych(seckey, msg, ring, base, exp)
		if err != nil {
			panic(err)
		}
		t1 := time.Now()
		fmt.Printf("Signing took %.3f sec\n", t1.Sub(t0).Seconds())

		// verify
		t2 := time.Now()
		ok := triptych.VerifyTriptych(sig, msg, ringUsed, base, exp)
		t3 := time.Now()
		if !ok {
			fmt.Println("Verification FAILED")
			os.Exit(1)
		}
		fmt.Printf("Verification OK. Took %.3f sec\n", t3.Sub(t2).Seconds())

		raw, _ := triptych.Serialize(sig)
		fmt.Printf("Raw signature length: %d bytes\n", len(raw))
		fmt.Printf("The signature's length is %d bytes, in base64, for %d keys.\n",
			len(base64.StdEncoding.EncodeToString(raw)), N)
		return
	}

	// If no args: print tiny usage and do a minimal quick self-test
	fmt.Println("Usage: go run ./cmd/test <n_base> <m_exp> <message>")
	fmt.Println("Example: go run ./cmd/test 3 3 \"hello\"")
}
