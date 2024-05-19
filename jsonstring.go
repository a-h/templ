package templ

import (
	"encoding/json"
)

// JSONString returns a JSON encoded string of v.
func JSONString(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
