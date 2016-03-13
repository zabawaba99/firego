package firego

import "encoding/json"

// Set the value of the Firebase reference.
func (fb *Firebase) Set(v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fb.doRequest("PUT", bytes)
	return err
}

func (fb *Firebase) SetValue(v interface{}) {
	fb.repo.SetValue(fb.path, v)
}
