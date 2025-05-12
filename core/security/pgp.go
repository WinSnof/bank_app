package security

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

func GenerateKeyPair(email string) (string, string, error) {
	// Генерация ключа
	entity, err := openpgp.NewEntity(email, "Generated PGP Key", "passphrase", nil)
	if err != nil {
		return "", "", err
	}

	// Экспорт открытого ключа
	var publicKeyBuf bytes.Buffer
	publicKeyWriter, err := armor.Encode(&publicKeyBuf, "PGP PUBLIC KEY BLOCK", nil)
	if err != nil {
		return "", "", err
	}
	defer publicKeyWriter.Close()

	if err := entity.Serialize(publicKeyWriter); err != nil {
		return "", "", err
	}

	// Экспорт закрытого ключа
	var privateKeyBuf bytes.Buffer
	privateKeyWriter, err := armor.Encode(&privateKeyBuf, "PGP PRIVATE KEY BLOCK", nil)
	if err != nil {
		return "", "", err
	}
	defer privateKeyWriter.Close()

	if err := entity.SerializePrivate(privateKeyWriter, nil); err != nil {
		return "", "", err
	}

	return publicKeyBuf.String(), privateKeyBuf.String(), nil
}

func MainGenerateKeyPair() {
	publicKeyFile := "public_key.asc"
	privateKeyFile := "private_key.asc"

	// Проверка существования ключей
	if _, err := os.Stat(publicKeyFile); err == nil {
		fmt.Println("Public key already exists. Skipping key generation.")
		return
	} else if !os.IsNotExist(err) {
		fmt.Println("Error checking public key file:", err)
		return
	}

	if _, err := os.Stat(privateKeyFile); err == nil {
		fmt.Println("Private key already exists. Skipping key generation.")
		return
	} else if !os.IsNotExist(err) {
		fmt.Println("Error checking private key file:", err)
		return
	}

	// Генерация ключа
	email := "user@example.com" // Замените на ваш email
	publicKey, privateKey, err := GenerateKeyPair(email)
	if err != nil {
		fmt.Println("Error generating key pair:", err)
		return
	}

	// Сохранение ключей в файлы
	if err := ioutil.WriteFile(publicKeyFile, []byte(publicKey), 0644); err != nil {
		fmt.Println("Error saving public key:", err)
		return
	}
	if err := ioutil.WriteFile(privateKeyFile, []byte(privateKey), 0644); err != nil {
		fmt.Println("Error saving private key:", err)
		return
	}

	fmt.Println("Public key:")
	fmt.Println(publicKey)
	fmt.Println("Private key:")
	fmt.Println(privateKey)
}
