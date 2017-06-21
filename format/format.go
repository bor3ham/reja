package format

type Page struct {
	Metadata map[string]interface{} `json:"meta"`
	Links    map[string]*string     `json:"links"`
	Data     []interface{}          `json:"data"`
	Included *[]interface{}         `json:"included,omitempty"`
}
