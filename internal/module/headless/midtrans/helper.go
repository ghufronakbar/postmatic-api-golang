package midtrans

import (
	"crypto/sha512"
	"encoding/hex"
)

// VerifySignature validates the signature from Midtrans notification callback
// Formula: SHA512(orderId + statusCode + grossAmount + serverKey)
func (s *midtransService) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	// Concatenate: orderID + statusCode + grossAmount + serverKey
	raw := orderID + statusCode + grossAmount + s.client.ServerKey

	// Hash using SHA512
	hasher := sha512.New()
	hasher.Write([]byte(raw))
	expected := hex.EncodeToString(hasher.Sum(nil))

	// Compare with provided signature
	return expected == signatureKey
}
