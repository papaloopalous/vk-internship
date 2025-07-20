package encryption

import (
	"api/internal/logger"
	"api/internal/messages"
	"api/internal/response"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"net/http"

	"github.com/spf13/viper"
)

// Криптографические параметры для обмена ключами по схеме Диффи-Хеллмана
var (
	strPrime   string
	prime      *big.Int
	generator  *big.Int
	serverPriv *big.Int
	serverPub  *big.Int
)

func init() {
	// Сначала инициализируем prime и generator
	strPrime = viper.GetString("crypto.prime")
	prime, _ = new(big.Int).SetString(strPrime, 16)
	generator = big.NewInt(viper.GetInt64("crypto.generator"))

	// Затем инициализируем серверные ключи
	serverPriv, _ = new(big.Int).SetString("1234567890ABCDEF1234567890ABCDEF12345678", 16)
	serverPub = new(big.Int).Exp(generator, serverPriv, prime)
}

// GetCryptoParams отправляет параметры для установки защищенного соединения
func GetCryptoParams(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		messages.CryptoParamPrime:     strPrime,
		messages.CryptoParamGenerator: generator.String(),
	}

	logger.Info(messages.ServiceEncryption, messages.LogStatusParamsSent, map[string]string{
		messages.LogPrime:     strPrime,
		messages.LogGenerator: generator.String(),
	})
	response.WriteAPIResponse(w, http.StatusOK, true, "", params)
}

// GetServerPublicKey возвращает публичный ключ сервера
func GetServerPublicKey() string {
	return serverPub.String()
}

// DeriveSharedKeyHex вычисляет общий секретный ключ по схеме Диффи-Хеллмана
func DeriveSharedKeyHex(clientPublic string) (string, error) {
	cliPub, ok := new(big.Int).SetString(clientPublic, 10)
	if !ok {
		logger.Error(messages.ServiceEncryption, messages.LogErrInvalidPublicKey, map[string]string{
			messages.LogKey: clientPublic,
		})
		return "", errors.New(messages.ClientErrInvalidPublicKey)
	}

	secret := new(big.Int).Exp(cliPub, serverPriv, prime)
	decStr := secret.String()
	hash := sha256.Sum256([]byte(decStr))

	logger.Info(messages.ServiceEncryption, messages.LogStatusKeyDerived, map[string]string{
		messages.LogKey: clientPublic,
	})
	return hex.EncodeToString(hash[:]), nil
}
