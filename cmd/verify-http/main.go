package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"coursach/triptych/triptych"
)

type VerifyRequest struct {
	Message      string   `json:"message"`
	SignatureB64 string   `json:"signatureB64"`
	Ring         []string `json:"ring"`
	N            int      `json:"n"`
	M            int      `json:"m"`
}

type VerifyResponse struct {
	OK      bool   `json:"ok"`
	UNumber string `json:"uNumber,omitempty"`
	Error   string `json:"error,omitempty"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/verify", handleVerify)

	addr := ":8088"
	log.Printf("verify-http listening on %s", addr)
	s := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}

func handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[verify] bad json: %v", err)
		writeJSON(w, http.StatusBadRequest, VerifyResponse{OK: false, Error: "bad json: " + err.Error()})
		return
	}

	log.Printf("[verify] new request: msg=%q n=%d m=%d ring=%d",
		req.Message, req.N, req.M, len(req.Ring))

	if req.N <= 1 || req.M <= 0 || len(req.Ring) == 0 || req.Message == "" || req.SignatureB64 == "" {
		log.Printf("[verify] missing fields")
		writeJSON(w, http.StatusBadRequest, VerifyResponse{OK: false, Error: "missing fields"})
		return
	}

	blob, err := base64.StdEncoding.DecodeString(req.SignatureB64)
	if err != nil || len(blob) < 33 {
		log.Printf("[verify] bad signature b64: %v", err)
		writeJSON(w, http.StatusBadRequest, VerifyResponse{OK: false, Error: "bad signature base64"})
		return
	}
	keyImg := blob[:33]
	raw := blob[33:]
	log.Printf("[verify] decoded signature: keyImg=%s raw=%d bytes",
		hex.EncodeToString(keyImg), len(raw))

	N := 1
	for i := 0; i < req.M; i++ {
		N *= req.N
	}
	if len(req.Ring) != N {
		log.Printf("[verify] ring length mismatch: got=%d expected=%d", len(req.Ring), N)
		writeJSON(w, http.StatusBadRequest, VerifyResponse{OK: false, Error: fmt.Sprintf("ring length must be n^m=%d", N)})
		return
	}
	ring := make([]*triptych.Point, N)
	for i, hx := range req.Ring {
		b, err := hex.DecodeString(hx)
		if err != nil {
			log.Printf("[verify] ring[%d] bad hex", i)
			writeJSON(w, http.StatusBadRequest, VerifyResponse{OK: false, Error: fmt.Sprintf("ring[%d] bad hex", i)})
			return
		}
		P, err := triptych.ParseCompressed(b)
		if err != nil {
			log.Printf("[verify] ring[%d] bad pubkey", i)
			writeJSON(w, http.StatusBadRequest, VerifyResponse{OK: false, Error: fmt.Sprintf("ring[%d] bad key", i)})
			return
		}
		ring[i] = P
	}

	sig, err := triptych.Deserialize(raw, req.M, req.N, keyImg)
	if err != nil {
		log.Printf("[verify] deserialize error: %v", err)
		writeJSON(w, http.StatusBadRequest, VerifyResponse{OK: false, Error: "deserialize: " + err.Error()})
		return
	}

	ok, uNumBytes := triptych.VerifyTriptych(sig, []byte(req.Message), ring, req.N, req.M)
	if !ok {
		log.Printf("[verify] signature invalid for msg=%s", req.Message)
		writeJSON(w, http.StatusOK, VerifyResponse{OK: false, Error: "invalid signature"})
		return
	}

	uNumHex := hex.EncodeToString(uNumBytes)
	elapsed := time.Since(start)
	log.Printf("[verify] signature OK for msg=%s uNum=%s (%.3fs)",
		req.Message, uNumHex, elapsed.Seconds())

	writeJSON(w, http.StatusOK, VerifyResponse{
		OK:      true,
		UNumber: uNumHex,
	})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
