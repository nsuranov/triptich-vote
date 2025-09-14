package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"coursach/triptych/triptych"
)

func readRing(path string) ([]*triptych.Point, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ring []*triptych.Point
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		b, err := hex.DecodeString(line)
		if err != nil {
			return nil, fmt.Errorf("bad hex pubkey: %w", err)
		}
		P, err := triptych.ParseCompressed(b)
		if err != nil {
			return nil, fmt.Errorf("bad pubkey: %w", err)
		}
		ring = append(ring, P)
	}
	return ring, sc.Err()
}

func main() {
	n := flag.Int("n", 3, "основание кольца")
	m := flag.Int("m", 3, "степень (размер кольца = n^m)")
	msg := flag.String("msg", "hello", "сообщение")
	sigB64 := flag.String("sig", "", "подпись (base64 от keyimg||raw)")
	ringFile := flag.String("ring", "", "файл с кольцом в порядке использования при подписи")
	flag.Parse()

	if *sigB64 == "" || *ringFile == "" {
		log.Fatalf("usage: verify -n 3 -m 3 -msg hi -sig sig.b64 -ring ring.used")
	}

	blob, err := base64.StdEncoding.DecodeString(*sigB64)
	if err != nil || len(blob) < 33 {
		log.Fatalf("bad base64: %v", err)
	}
	keyImg := blob[:33]
	raw := blob[33:]

	ring, err := readRing(*ringFile)
	if err != nil {
		log.Fatalf("read ring: %v", err)
	}

	sig, err := triptych.Deserialize(raw, *m, *n, keyImg)
	if err != nil {
		log.Fatalf("deserialize: %v", err)
	}

	ok, _ := triptych.VerifyTriptych(sig, []byte(*msg), ring, *n, *m)
	if !ok {
		fmt.Println("Verification FAILED")
		os.Exit(1)
	}
	fmt.Println("Verification OK")
}
