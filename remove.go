package firego

// Remove the Firebase reference from the cloud
func (fb *Firebase) Remove() error {
	_, err := fb.doRequest("DELETE", nil)
	if err != nil {
		return err
	}
	return nil
}
