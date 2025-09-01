package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"coursach/triptych/triptych"
)

// ===== DTOs с бэкенда =====

type CandidateDTO struct {
	ID       string `json:"id"`
	Fullname string `json:"fullname"`
}

type RingDTO struct {
	PublicKeys []string `json:"publicKeys"` // 33B compressed hex
}

// что отправляем на сервер бюллетеня
type BulletinCreateDTO struct {
	CandidateID  string   `json:"candidateId"`
	SignatureB64 string   `json:"signatureB64"` // base64(keyImage||rawSig)
	Ring         []string `json:"ring"`         // порядок кольца, использованный при подписи (33B hex)
	N            int      `json:"n"`
	M            int      `json:"m"`
}

// файл ключей, как мы делали в keygen
type keypairFile struct {
	FullName  string `json:"fullName"`
	PublicKey string `json:"publicKey"` // 33B hex
	SecretKey string `json:"secretKey"` // 32B hex
	CreatedAt string `json:"createdAt"`
}

// ===== main =====

func main() {
	baseURL := flag.String("url", "", "базовый URL бэкенда (например http://localhost:8080)")
	keysPath := flag.String("keys", "", "путь к файлу с парой ключей (JSON из keygen)")
	n := flag.Int("n", 3, "основание кольца (n)")
	flag.Parse()

	if *baseURL == "" || *keysPath == "" {
		log.Fatalf("usage: vote -url http://localhost:8080 -keys ./alice-key.json [-n 3]")
	}

	// 1) читаем ключи
	kf, sk, _ := loadKeys(*keysPath)

	// 2) тянем кандидатов
	cands := fetchCandidates(*baseURL)
	if len(cands) == 0 {
		log.Fatal("список кандидатов пуст")
	}

	fmt.Println("Кандидаты:")
	for i, c := range cands {
		fmt.Printf("[%d] %s  (%s)\n", i, c.Fullname, c.ID)
	}
	idx := askIndex(len(cands))
	cand := cands[idx]
	msg := []byte(cand.ID) // подписываем UUID кандидата

	// 3) тянем кольцо
	ringPoints, ringHex := fetchRing(*baseURL)

	// проверяем, что наш pk в кольце
	if !containsHex(ringHex, kf.PublicKey) {
		log.Fatalf("ваш публичный ключ отсутствует в кольце. Сначала зарегистрируйте его через keygen, затем повторите попытку")
	}

	// 4) приводим размер кольца к n^m: подбираем m и при необходимости дополняем фиктивными ключами
	ringPoints, m, targetN := chooseMAndPadRing(*n, ringPoints)
	sig, ringUsed, err := triptych.RingSignTriptych(sk, msg, ringPoints, *n, m)

	if err != nil {
		log.Fatalf("sign: %v", err)
	}
	raw, keyImg := triptych.Serialize(sig)

	// 6) подготавливаем и отправляем бюллетень
	sigB64 := base64.StdEncoding.EncodeToString(append(keyImg, raw...))
	ringUsedHex := make([]string, len(ringUsed))
	for i, p := range ringUsed {
		ringUsedHex[i] = hex.EncodeToString(p.BytesCompressed())
	}

	payload := BulletinCreateDTO{
		CandidateID:  cand.ID,
		SignatureB64: sigB64,
		Ring:         ringUsedHex,
		N:            *n,
		M:            m,
	}

	sendBulletin(*baseURL, payload)

	fmt.Printf("\nГолос отправлен. Параметры: n=%d, m=%d (ring=%d, n^m=%d)\n", *n, m, len(ringUsedHex), targetN)
	fmt.Printf("Ваш uNumber (key image): %s\n", hex.EncodeToString(keyImg))
	fmt.Println("Важно: храните uNumber — по нему можно обнаружить повторный голос.")
}

// ===== helpers: io/network =====

func loadKeys(path string) (keypairFile, []byte, *triptych.Point) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read keys: %v", err)
	}
	var kf keypairFile
	if err := json.Unmarshal(b, &kf); err != nil {
		log.Fatalf("parse keys json: %v", err)
	}
	if len(kf.SecretKey) != 64 || len(kf.PublicKey) != 66 {
		log.Fatalf("неверный формат ключей (ожидается hex: sk=32B, pk=33B)")
	}
	sk, err := hex.DecodeString(kf.SecretKey)
	if err != nil {
		log.Fatalf("bad sk hex: %v", err)
	}
	pkb, _ := hex.DecodeString(kf.PublicKey)
	pk, err := triptych.ParseCompressed(pkb)
	if err != nil {
		log.Fatalf("bad pk: %v", err)
	}
	return kf, sk, pk
}

func fetchCandidates(baseURL string) []CandidateDTO {
	u := strings.TrimRight(baseURL, "/") + "/api/candidate"
	var cands []CandidateDTO
	if err := doJSON(http.MethodGet, u, nil, &cands); err != nil {
		log.Fatalf("fetch candidates: %v", err)
	}
	return cands
}

func fetchRing(baseURL string) ([]*triptych.Point, []string) {
	u := strings.TrimRight(baseURL, "/") + "/api/signer/ring"
	var resp RingDTO
	if err := doJSON(http.MethodGet, u, nil, &resp); err != nil {
		log.Fatalf("fetch ring: %v", err)
	}
	if len(resp.PublicKeys) == 0 {
		log.Fatalf("получено пустое кольцо")
	}
	var ring []*triptych.Point
	for i, hx := range resp.PublicKeys {
		b, err := hex.DecodeString(hx)
		if err != nil {
			log.Fatalf("ring[%d] bad hex: %v", i, err)
		}
		P, err := triptych.ParseCompressed(b)
		if err != nil {
			log.Fatalf("ring[%d] bad pubkey: %v", i, err)
		}
		ring = append(ring, P)
	}
	return ring, resp.PublicKeys
}

func sendBulletin(baseURL string, payload BulletinCreateDTO) {
	u := strings.TrimRight(baseURL, "/") + "/api/bulletin"
	if err := doJSON(http.MethodPost, u, payload, nil); err != nil {
		log.Fatalf("send bulletin: %v", err)
	}
	fmt.Println("Сервер принял бюллетень (HTTP 2xx).")
}

func doJSON(method, url string, in interface{}, out interface{}) error {
	var body *bytes.Reader
	if in != nil {
		b, _ := json.Marshal(in)
		body = bytes.NewReader(b)
	} else {
		body = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("http %d", res.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}
	return nil
}

// ===== helpers: ring math / ui =====

func askIndex(max int) int {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Выберите номер кандидата [0..%d]: ", max-1)
		line, _ := r.ReadString('\n')
		line = strings.TrimSpace(line)
		i, err := strconv.Atoi(line)
		if err == nil && i >= 0 && i < max {
			return i
		}
		fmt.Println("Неверный ввод")
	}
}

func containsHex(arr []string, needle string) bool {
	needle = strings.ToLower(needle)
	for _, s := range arr {
		if strings.ToLower(s) == needle {
			return true
		}
	}
	return false
}

func chooseMAndPadRing(n int, ring []*triptych.Point) (padded []*triptych.Point, m int, targetN int) {
	L := len(ring)

	// подобрать минимальное m такое, что n^m >= L
	targetN = 1
	m = 0
	for targetN < L {
		targetN *= n
		m++
	}
	if m == 0 {
		m = 1
		targetN = n
	}

	// дополняем фиктивными ключами до targetN
	padded = ring
	for len(padded) < targetN {
		_, pk := triptych.GenerateKey()
		padded = append(padded, pk)
	}
	return padded, m, targetN
}
