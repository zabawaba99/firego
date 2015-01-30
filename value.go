package firego

import "encoding/json"

// Value gets the value of the Firebase reference
func (fb *Firebase) Value(v interface{}) error {
	bytes, err := fb.doRequest("GET", nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}
