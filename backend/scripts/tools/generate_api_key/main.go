package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/tools/generate_api_key.go <client_name>")
		os.Exit(1)
	}

	clientName := os.Args[1]

	// generate random key, prefix biar gampang dikenali sumbernya
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		fmt.Println("Error generate random:", err)
		os.Exit(1)
	}
	apiKey := "bdk_" + hex.EncodeToString(randomBytes) // bdk = bank-data key

	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing:", err)
		os.Exit(1)
	}

	fmt.Println("=================================================")
	fmt.Println("Client Name :", clientName)
	fmt.Println("API Key     :", apiKey)
	fmt.Println("Hash (utk DB):", string(hash))
	fmt.Println("=================================================")
	fmt.Println("⚠️  Simpan API Key di atas sekarang — tidak akan ditampilkan lagi!")
	fmt.Println("⚠️  Yang disimpan ke database adalah HASH-nya, bukan key asli.")
}
