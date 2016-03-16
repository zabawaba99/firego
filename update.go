package firego

import "encoding/json"

// Update the specific child with the given value.
func (fb *Firebase) Update(v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fb.doRequest("PATCH", bytes)
	return err
}

func (fb *Firebase) UpdateChildren(children map[string]interface{}) {
	fb.repo.UpdateChildren(fb.path, children)
}
