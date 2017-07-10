package schema

import (
	"encoding/json"
)

type Result struct {
	Provided bool `json:"-"`

	Links    map[string]*string `json:"links"`
	Data     interface{}        `json:"data"`
	Included *[]interface{}     `json:"included,omitempty"`
}

func (r *Result) UnmarshalJSON(data []byte) error {
	r.Provided = true
	var val struct {
		Links    map[string]*string `json:"links"`
		Data     interface{}        `json:"data"`
		Included *[]interface{}     `json:"included,omitempty"`
	}
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	r.Links = val.Links
	r.Data = val.Data
	r.Included = val.Included
	return nil
}
