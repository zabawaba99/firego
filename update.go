package firego

import "encoding/json"

// Update the specific child with the given value
func (fb *Firebase) Update(v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fb.doRequest("PATCH", bytes)
	return err
}
