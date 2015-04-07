package firego

import "encoding/json"

// Set the value of the Firebase reference
func (fb *Firebase) Set(v interface{}, params map[string]string) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fb.doRequest("PUT", params, bytes)
	return err
}
