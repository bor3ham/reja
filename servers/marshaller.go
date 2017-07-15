package servers

import (
	"bytes"
	"encoding/json"
)

func MustJSONMarshal(v interface{}) []byte {
	// b, err := json.MarshalIndent(v, "", "    ")
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
	b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
	b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	return b
}

func JSONUnmarshal(b []byte, shape interface{}) error {
	return json.Unmarshal(b, shape)
}
