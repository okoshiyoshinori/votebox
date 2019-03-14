package util

import (
	"github.com/okoshiyoshinori/votebox/config"
	"github.com/speps/go-hashids"
)

var (
	h *hashids.HashID
)

func init() {
	hd := hashids.NewData()
	hd.Salt = config.APConfig.Salt
	hd.MinLength = 8
	h, _ = hashids.NewWithData(hd)
}

func Encode(s int) string {
	e, _ := h.Encode([]int{s})
	return e
}

func Decode(s string) (int, error) {
	d, err := h.DecodeWithError(s)
	if err != nil {
		return 0, err
	}
	return d[0], nil
}
