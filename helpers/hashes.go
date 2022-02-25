package helpers

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/speps/go-hashids/v2"
)

func GenerateMetadataHashID(uuidString string, tokenID int, debugPrint bool) (string, error) {
	hd := hashids.NewData()
	hd.Salt = uuidString
	hd.MinLength = 10
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return  "", terror.Error(err)
	}

	e, err := h.Encode([]int{tokenID})
	if err != nil {
		return  "", terror.Error(err)
	}
	d, err := h.DecodeWithError(e)
	if err != nil {
		return  "", terror.Error(err)
	}

	if debugPrint {
		fmt.Printf("Hash for %s token id %d. Hash: %s, Decoded: %v\n", hd.Salt,  tokenID, e, d)
	}

	return e, nil
}
