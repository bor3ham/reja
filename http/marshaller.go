package http

import (
	"bytes"
	"encoding/json"
)

func MustJSONMarshal(v interface{}) []byte {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
	b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
	b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	return b
}

func MustJSONUnmarshal(b []byte, shape interface{}) {
	err := json.Unmarshal(b, shape)
	if err != nil {
		panic(err)
	}
}
