package main

import (
	"encoding/hex"
	"strings"
)

type address []byte

func (b address) Pretty() string {
	return "0x" + hex.EncodeToString(b)
}

type addressSet []string

func (s addressSet) contains(address string) bool {
	for _, candidate := range s {
		if strings.EqualFold(candidate, address) {
			return true
		}
	}

	return false
}
