package firego

import "encoding/json"

// Push creates a reference to an auto-generated child location
func (fb *Firebase) Push(v interface{}, params map[string]string) (*Firebase, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	bytes, err = fb.doRequest("POST", params, bytes)
	var m map[string]string
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}
	return &Firebase{
		url:    fb.url + "/" + m["name"],
		client: fb.client,
	}, err
}
