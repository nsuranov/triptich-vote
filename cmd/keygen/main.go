package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"coursach/triptych/triptych"
)

type signerCreateDTO struct {
	FullName  string `json:"fullName"`
	PublicKey string `json:"publicKey"`
}

type keypairFile struct {
	FullName  string `json:"fullName"`
	PublicKey string `json:"publicKey"` // 33B compressed, hex
	SecretKey string `json:"secretKey"` // 32B hex
	CreatedAt string `json:"createdAt"`
}

func main() {
	// optional flags
	outPath := flag.String("out", "", "путь к файлу для сохранения пары ключей (по умолчанию: <name>-key.json)")
	flag.Parse()

	// required positional args: <fullName> <baseURL>
	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("usage: keygen [-out file.json] <fullName> <baseURL>")
		fmt.Println("example: keygen \"Alice Smith\" http://localhost:8080")
		os.Exit(2)
	}
	fullName := args[0]
	baseURL := args[1]

	// 1) generate keys
	sk, pk := triptych.GenerateKey()
	pkHex := hex.EncodeToString(pk.BytesCompressed())
	skHex := hex.EncodeToString(sk)

	// 2) save to file (0600)
	fileName := *outPath
	if fileName == "" {
		fileName = defaultFileName(fullName)
	}
	kf := keypairFile{
		FullName:  fullName,
		PublicKey: pkHex,
		SecretKey: skHex,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	if err := writeKeyFile(fileName, kf); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write key file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Keypair saved to %s\n", fileName)

	// 3) HTTP POST to <baseURL>/api/signer
	registerURL := strings.TrimRight(baseURL, "/") + "/api/signer"
	payload := signerCreateDTO{FullName: fullName, PublicKey: pkHex}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, registerURL, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "registration request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "registration failed: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Printf("Registered successfully at %s\n", registerURL)
}

// --- helpers ---

func writeKeyFile(path string, kf keypairFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		// ignore error if path has no dir component
	}
	b, _ := json.MarshalIndent(kf, "", "  ")
	// 0600: чтобы приватный ключ не был доступен другим пользователям
	return os.WriteFile(path, b, 0o600)
}

func defaultFileName(fullName string) string {
	s := strings.ToLower(strings.TrimSpace(fullName))
	// заменить всё, кроме латиницы/цифр/дефиса/подчёркивания, на "_"
	re := regexp.MustCompile(`[^a-z0-9\-_]+`)
	s = re.ReplaceAllString(s, "_")
	if s == "" {
		s = "signer"
	}
	return s + "-key.json"
}
