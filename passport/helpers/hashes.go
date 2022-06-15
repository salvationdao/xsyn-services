package helpers

import (
	"fmt"

	"github.com/speps/go-hashids/v2"
)

func GenerateMetadataHashID(uuidString string, tokenID int, debugPrint bool) (string, error) {
	hd := hashids.NewData()
	hd.Salt = uuidString
	hd.MinLength = 10
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}

	e, err := h.Encode([]int{tokenID})
	if err != nil {
		return "", err
	}
	d, err := h.DecodeWithError(e)
	if err != nil {
		return "", err
	}

	if debugPrint {
		fmt.Printf("Hash for %s token id %d. Hash: %s, Decoded: %v\n", hd.Salt, tokenID, e, d)
	}

	return e, nil
}

// UNRELIABLE DUE TO DIRTY DATA, DO NOT USE
// Or at least recalculate the hashes first
func UnhashMetadataHashID(collectionUUID, hash string) (int, error) {
	hd := hashids.NewData()
	hd.Salt = collectionUUID
	hd.MinLength = 10
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return 0, err
	}

	d, err := h.DecodeWithError(hash)
	if err != nil {
		return 0, err
	}

	return d[0], nil
}
