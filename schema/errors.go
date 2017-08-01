package schema

type ErrorSet struct {
	Errors []map[string]interface{} `json:"errors"`
}
