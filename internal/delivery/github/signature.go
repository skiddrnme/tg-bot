package github

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

func VerifySignature(payload []byte, header, secret string) bool {
    const prefix = "sha256="
    if !strings.HasPrefix(header, prefix) {
        return false
    }
    gotMAC, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
    if err != nil {
        return false
    }
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    return hmac.Equal(gotMAC, mac.Sum(nil))
}