package format

import (
	"encoding/json"
)

type Page struct {
	Provided bool `json:"-"`

	Metadata map[string]interface{} `json:"meta"`
	Links    map[string]*string     `json:"links"`
	Data     []interface{}          `json:"data"`
	Included *[]interface{}         `json:"included,omitempty"`
}

func (p *Page) UnmarshalJSON(data []byte) error {
	p.Provided = true
	var val struct {
		Metadata map[string]interface{} `json:"meta"`
		Links    map[string]*string     `json:"links"`
		Data     []interface{}          `json:"data"`
		Included *[]interface{}         `json:"included,omitempty"`
	}
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	p.Metadata = val.Metadata
	p.Links = val.Links
	p.Data = val.Data
	p.Included = val.Included
	return nil
}
