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

func writeRing(path string, ring []*triptych.Point) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, p := range ring {
		fmt.Fprintln(w, hex.EncodeToString(p.BytesCompressed()))
	}
	return w.Flush()
}

func main() {
	n := flag.Int("n", 3, "основание кольца")
	m := flag.Int("m", 3, "степень (размер кольца = n^m)")
	msg := flag.String("msg", "hello", "сообщение для подписи")
	skHex := flag.String("sk", "", "секретный ключ (32B hex)")
	ringFile := flag.String("ring", "", "файл со списком публичных ключей (по одному 33B hex в строке)")
	outSig := flag.String("out", "sig.b64", "куда сохранить подпись (base64)")
	outRing := flag.String("out-ring", "ring.used", "куда сохранить порядок кольца, использованный при подписи")
	flag.Parse()

	if *skHex == "" || *ringFile == "" {
		log.Fatalf("usage: sign -n 3 -m 3 -msg \"hi\" -sk <hex> -ring ring.txt")
	}

	sk, err := hex.DecodeString(*skHex)
	if err != nil || len(sk) != 32 {
		log.Fatalf("bad sk hex: %v", err)
	}
	ring, err := readRing(*ringFile)
	if err != nil {
		log.Fatalf("read ring: %v", err)
	}

	sig, ringUsed, err := triptych.RingSignTriptych(sk, []byte(*msg), ring, *n, *m)
	if err != nil {
		log.Fatalf("sign: %v", err)
	}

	raw, keyImg := triptych.Serialize(sig)

	if err := os.WriteFile(*outSig, []byte(base64.StdEncoding.EncodeToString(append(keyImg, raw...))), 0644); err != nil {
		log.Fatalf("write sig: %v", err)
	}
	if err := writeRing(*outRing, ringUsed); err != nil {
		log.Fatalf("write ring: %v", err)
	}

	fmt.Printf("OK. Signature saved to %s (base64 of keyimage||raw). Ring order to %s.\n", *outSig, *outRing)
	fmt.Printf("Raw sig bytes: %d\n", len(raw))
}
