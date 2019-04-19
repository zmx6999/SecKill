package service

import (
	"math/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateNonceStr(length int) string {
	x := "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	xBytes := []byte(x)
	str := ""
	for i := 0; i < length; i++ {
		str += string(xBytes[rand.Intn(len(xBytes))])
	}
	return str
}

func GetSha256(x string) string {
	h := sha256.New()
	h.Write([]byte(x))
	return hex.EncodeToString(h.Sum(nil))
}
