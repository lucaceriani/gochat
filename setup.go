package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"os"
	"path"
	"strings"
	"syscall"

	"golang.org/x/crypto/sha3"
	"golang.org/x/term"
)

func getKeyPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return path.Join(homedir, ".gochat.key")
}

func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	// error in decryption
	if unpadding > length {
		return src
	}
	return src[:(length - unpadding)]
}

func decryptKey(apiKeyEnc []byte, iv []byte, pass string, checksum string) (string, error) {
	h := sha3.New256()
	h.Write([]byte(pass))
	passwordHash := h.Sum(nil)

	// 3.3. decrypt the api key with the password hash and the IV
	block, _ := aes.NewCipher(passwordHash)
	mode := cipher.NewCBCDecrypter(block, iv)

	apiKey := make([]byte, len(apiKeyEnc))
	mode.CryptBlocks(apiKey, apiKeyEnc)

	// 3.4. unpad the api key
	apiKey = PKCS5UnPadding(apiKey)

	// 3.5. check the checksum
	decriptionChecksumStr := fmt.Sprintf("%x", crc32.ChecksumIEEE(apiKey))

	if decriptionChecksumStr != checksum {
		return "", fmt.Errorf("invalid password")
	} else {
		return string(apiKey), nil
	}
}

func getKey() (string, error) {
	keyFileContents, err := os.ReadFile(getKeyPath())

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Key file not found!\nPlease run gochat setup with \"gochat setup\"")
			os.Exit(1)
		} else {
			panic(err)
		}
	}

	// decrypt the key with the password
	splitted := strings.Split(string(keyFileContents), "\n")
	if len(splitted) < 3 {
		panic("invalid key file")
	}

	iv, _ := base64.StdEncoding.DecodeString(splitted[0])
	apiKeyEnc, _ := base64.StdEncoding.DecodeString(splitted[1])
	checksumStr := splitted[2]

	// try to decrypt the key with an empty password
	apiKey, err := decryptKey(apiKeyEnc, iv, "", checksumStr)
	if err == nil {
		return apiKey, nil
	}

	// ask the user for the password
	fmt.Print("Password for decryption: ")
	bytePass, _ := term.ReadPassword(int(syscall.Stdin))
	password := strings.TrimSpace(string(bytePass))

	return decryptKey(apiKeyEnc, iv, password, checksumStr)

}

// setup function, it has to
func setup() {
	fmt.Print("GoChat initial setup\n\n")

	// 1. ask the user for the api key
	fmt.Println("Please enter your API key: ")
	fmt.Print("> ")
	apiKey, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	// 2. ask the user for the password
	fmt.Println("Please enter a password for encryption: (can be empty)")
	fmt.Print("> ")
	bytePass, _ := term.ReadPassword(int(syscall.Stdin))
	password := strings.TrimSpace(string(bytePass))

	if len(password) != 0 {
		fmt.Println("\nPlease enter the password again: (can be empty)")
		fmt.Print("> ")
		bytePassCheck, _ := term.ReadPassword(int(syscall.Stdin))
		passwordCheck := strings.TrimSpace(string(bytePassCheck))

		if passwordCheck != password {
			fmt.Println("\nPasswords do not match! D:")
			os.Exit(1)
		}
	}

	// 3. encrypt the api key with the password using AES CBC and for the IV a random 16 bytes string

	// 3.1. hash the password with sha3-256
	h := sha3.New256()
	h.Write([]byte(password))
	passwordHash := h.Sum(nil)

	// 3.2. generate a random 128b string for the IV
	iv := make([]byte, 128/8)
	rand.Read(iv)

	// 3.3 pad the api key to a multiple of 128b
	apiKeyPad := PKCS5Padding([]byte(apiKey), 128/8, len(apiKey))

	// 3.3. encrypt the api key with the password hash and the IV
	block, _ := aes.NewCipher(passwordHash)
	mode := cipher.NewCBCEncrypter(block, iv)

	apiKeyEnc := make([]byte, len(apiKeyPad))
	mode.CryptBlocks(apiKeyEnc, apiKeyPad)

	// 4. save the iv and the encrypted api key to a file with base64 encoding

	apiKeyEncB64 := base64.StdEncoding.EncodeToString(apiKeyEnc)
	ivB64 := base64.StdEncoding.EncodeToString(iv)

	// 5. get the crc32 of the api key and save it to the file
	checksum := fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(apiKey)))

	os.WriteFile(getKeyPath(), []byte(ivB64+"\n"+apiKeyEncB64+"\n"+checksum), 0644)

}
