package firego

// Remove the Firebase reference from the cloud
func (fb *Firebase) Remove(params map[string]string) error {
	_, err := fb.doRequest("DELETE", params, nil)
	if err != nil {
		return err
	}
	return nil
}
