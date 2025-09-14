package main

import (
	"coursach/triptych/triptych"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

// ---- Конфигурация запуска ----

type BenchConfig struct {
	Base    int    `json:"base"`     // всегда 2
	MinExp  int    `json:"min_exp"`  // по умолчанию 0
	MaxExp  int    `json:"max_exp"`  // по умолчанию 12
	Trials  int    `json:"trials"`   // по умолчанию 3
	Message string `json:"message"`  // по умолчанию "hello"
	OutPath string `json:"out_path"` // по умолчанию triptych_bench_results.json
}

// ---- Результаты ----

type TrialResult struct {
	Exp            int     `json:"exp"`
	RingSize       int     `json:"ring_size"`
	Trial          int     `json:"trial"`
	SignMS         float64 `json:"sign_ms"`
	VerifyMS       float64 `json:"verify_ms"`       // только VerifyTriptych
	VerifyTotalMS  float64 `json:"verify_total_ms"` // Deserialize + VerifyTriptych
	SigLenBytes    int     `json:"sig_len_bytes"`   // keyimage + raw
	RawLenBytes    int     `json:"raw_len_bytes"`
	KeyImageBytes  int     `json:"keyimage_bytes"`
	MessageLenByte int     `json:"message_len_byte"`
}

type Aggregate struct {
	Exp           int     `json:"exp"`
	RingSize      int     `json:"ring_size"`
	Trials        int     `json:"trials"`
	SignAvgMS     float64 `json:"sign_avg_ms"`
	SignMinMS     float64 `json:"sign_min_ms"`
	SignMaxMS     float64 `json:"sign_max_ms"`
	VerifyAvgMS   float64 `json:"verify_avg_ms"`
	VerifyMinMS   float64 `json:"verify_min_ms"`
	VerifyMaxMS   float64 `json:"verify_max_ms"`
	VerifyTotAvg  float64 `json:"verify_total_avg_ms"`
	SigLenAvg     float64 `json:"sig_len_avg_bytes"`
	RawLenAvg     float64 `json:"raw_len_avg_bytes"`
	KeyImageBytes int     `json:"keyimage_bytes"`
}

type BenchOutput struct {
	Config    BenchConfig   `json:"config"`
	Timestamp string        `json:"timestamp"`
	Results   []Aggregate   `json:"results"`
	Trials    []TrialResult `json:"trials"`
}

// ---- Вспомогательные функции ----

func powInt(base, exp int) int {
	res := 1
	for i := 0; i < exp; i++ {
		res *= base
	}
	return res
}

func minMaxAvg(vals []float64) (min, max, avg float64) {
	if len(vals) == 0 {
		return 0, 0, 0
	}
	min, max, sum := vals[0], vals[0], 0.0
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	avg = sum / float64(len(vals))
	return
}

// генерирует кольцо размера ringSize; первый элемент — pk сигнера.
// Остальные — случайные публичные ключи (уникальные), полученные из GenerateKey().
func buildRingWithSignerFirst(pkCompressed []byte, ringSize int) ([]*triptych.Point, error) {
	if ringSize <= 0 {
		return nil, fmt.Errorf("ringSize must be > 0")
	}
	// Парсим публиковый ключ сигнера в Point
	signerPoint, err := triptych.ParseCompressed(pkCompressed)
	if err != nil {
		return nil, fmt.Errorf("parse signer pubkey: %w", err)
	}
	ring := make([]*triptych.Point, 0, ringSize)
	ring = append(ring, signerPoint)

	// Чтобы избежать дубликатов, сохраняем hex сжатых ключей
	seen := map[string]struct{}{
		hex.EncodeToString(pkCompressed): {},
	}

	for len(ring) < ringSize {
		_, pk := triptych.GenerateKey()
		b := pk.BytesCompressed()
		h := hex.EncodeToString(b)
		if _, ok := seen[h]; ok {
			continue
		}
		seen[h] = struct{}{}
		p, err := triptych.ParseCompressed(b)
		if err != nil {
			return nil, fmt.Errorf("parse generated pubkey: %w", err)
		}
		ring = append(ring, p)
	}

	return ring, nil
}

func main() {
	// --- Флаги ---
	outPath := flag.String("out", "triptych_bench_results.json", "путь к JSON с результатами")
	maxExp := flag.Int("max-exp", 15, "максимальная степень exp (размер кольца = 2^exp)")
	trials := flag.Int("trials", 1, "число повторов на каждую конфигурацию")
	msg := flag.String("msg", "d2c51a8e-344d-4f76-8458-119e4fb077a", "сообщение для подписи")
	flag.Parse()

	cfg := BenchConfig{
		Base:    2,
		MinExp:  1,
		MaxExp:  *maxExp,
		Trials:  *trials,
		Message: *msg,
		OutPath: *outPath,
	}

	if cfg.MaxExp < cfg.MinExp {
		log.Fatalf("max-exp (%d) must be >= min-exp (%d)", cfg.MaxExp, cfg.MinExp)
	}
	if cfg.Trials <= 0 {
		log.Fatalf("trials must be > 0")
	}

	log.Printf("Triptych benchmark starting...")
	log.Printf("Base=%d, exp in [%d..%d], trials=%d, msg=%q\n", cfg.Base, cfg.MinExp, cfg.MaxExp, cfg.Trials, cfg.Message)

	// 1) Генерируем одну пару ключей для подписанта
	sk, pk := triptych.GenerateKey()
	pkCompressed := pk.BytesCompressed()
	log.Printf("Signer key generated. Pub (compressed hex)=%s\n", hex.EncodeToString(pkCompressed))

	allTrials := make([]TrialResult, 0)
	aggs := make([]Aggregate, 0)

	msgBytes := []byte(cfg.Message)

	for exp := cfg.MinExp; exp <= cfg.MaxExp; exp++ {
		ringSize := powInt(cfg.Base, exp)

		// 2) Формируем кольцо указанного размера, включая наш pk
		ring, err := buildRingWithSignerFirst(pkCompressed, ringSize)
		if err != nil {
			log.Fatalf("build ring (exp=%d, size=%d): %v", exp, ringSize, err)
		}

		signTimes := make([]float64, 0, cfg.Trials)
		verifyTimes := make([]float64, 0, cfg.Trials)
		verifyTotalTimes := make([]float64, 0, cfg.Trials)
		sigLens := make([]float64, 0, cfg.Trials)
		rawLens := make([]float64, 0, cfg.Trials)
		keyImgLen := 0

		log.Printf("=== exp=%d, ringSize=%d ===", exp, ringSize)
		for t := 1; t <= cfg.Trials; t++ {
			// 3) Подпись
			t0 := time.Now()
			sig, ringUsed, err := triptych.RingSignTriptych(sk, msgBytes, ring, cfg.Base, exp)
			if err != nil {
				log.Fatalf("sign failed (exp=%d, trial=%d): %v", exp, t, err)
			}
			signDur := time.Since(t0)

			// сериализация
			raw, keyImage := triptych.Serialize(sig)
			totalSigLen := len(raw) + len(keyImage)
			keyImgLen = len(keyImage) // одинаковый для всех прогонов в рамках exp

			// 4) Проверка — меряем отдельно чистое VerifyTriptych и «полное» (Deserialize + Verify)
			t1 := time.Now()
			sig2, err := triptych.Deserialize(raw, exp, cfg.Base, keyImage)
			if err != nil {
				log.Fatalf("deserialize failed (exp=%d, trial=%d): %v", exp, t, err)
			}
			afterDeserialize := time.Now()
			ok, _ := triptych.VerifyTriptych(sig2, msgBytes, ringUsed, cfg.Base, exp)
			//if err != nil {
			//	log.Fatalf("verify error (exp=%d, trial=%d): %v", exp, t, err)
			//}
			if !ok {
				log.Fatalf("verify failed (exp=%d, trial=%d)", exp, t)
			}
			verifyTotalDur := time.Since(t1)
			verifyPureDur := time.Since(afterDeserialize)

			tr := TrialResult{
				Exp:            exp,
				RingSize:       ringSize,
				Trial:          t,
				SignMS:         float64(signDur.Microseconds()) / 1000.0,
				VerifyMS:       float64(verifyPureDur.Microseconds()) / 1000.0,
				VerifyTotalMS:  float64(verifyTotalDur.Microseconds()) / 1000.0,
				SigLenBytes:    totalSigLen,
				RawLenBytes:    len(raw),
				KeyImageBytes:  keyImgLen,
				MessageLenByte: len(msgBytes),
			}
			allTrials = append(allTrials, tr)

			signTimes = append(signTimes, tr.SignMS)
			verifyTimes = append(verifyTimes, tr.VerifyMS)
			verifyTotalTimes = append(verifyTotalTimes, tr.VerifyTotalMS)
			sigLens = append(sigLens, float64(tr.SigLenBytes))
			rawLens = append(rawLens, float64(tr.RawLenBytes))

			log.Printf("[exp=%d, ring=%d] trial=%d  sign=%.3f ms | verify=%.3f ms (total=%.3f ms) | sig_len=%d bytes (raw=%d, keyimg=%d)",
				exp, ringSize, t, tr.SignMS, tr.VerifyMS, tr.VerifyTotalMS, tr.SigLenBytes, tr.RawLenBytes, tr.KeyImageBytes)
		}

		sMin, sMax, sAvg := minMaxAvg(signTimes)
		vMin, vMax, vAvg := minMaxAvg(verifyTimes)
		_, _, vTotAvg := minMaxAvg(verifyTotalTimes)
		_, _, sigAvg := minMaxAvg(sigLens)
		_, _, rawAvg := minMaxAvg(rawLens)

		ag := Aggregate{
			Exp:           exp,
			RingSize:      ringSize,
			Trials:        cfg.Trials,
			SignAvgMS:     sAvg,
			SignMinMS:     sMin,
			SignMaxMS:     sMax,
			VerifyAvgMS:   vAvg,
			VerifyMinMS:   vMin,
			VerifyMaxMS:   vMax,
			VerifyTotAvg:  vTotAvg,
			SigLenAvg:     sigAvg,
			RawLenAvg:     rawAvg,
			KeyImageBytes: keyImgLen,
		}
		aggs = append(aggs, ag)

		log.Printf(">>> [exp=%d, ring=%d] SIGN avg=%.3f ms (min=%.3f, max=%.3f) | VERIFY avg=%.3f ms (min=%.3f, max=%.3f) | VERIFY(total) avg=%.3f ms | SIG avg≈%.0f bytes (raw≈%.0f, keyimg=%d)",
			exp, ringSize, sAvg, sMin, sMax, vAvg, vMin, vMax, vTotAvg, sigAvg, rawAvg, keyImgLen)
	}

	out := BenchOutput{
		Config:    cfg,
		Timestamp: time.Now().Format(time.RFC3339),
		Results:   aggs,
		Trials:    allTrials,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatalf("marshal results: %v", err)
	}
	if err := os.WriteFile(cfg.OutPath, data, 0644); err != nil {
		log.Fatalf("write %s: %v", cfg.OutPath, err)
	}
	log.Printf("JSON results saved to %s", cfg.OutPath)
	log.Printf("Done.")
}
