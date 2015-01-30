package firego

// SetAuth sets the custom Firebase token used to authenticate to Firebase
func (fb *Firebase) SetAuth(token string) {
	fb.auth = token
}

// RemoveAuth removes the current token being used to authenticate to Firebase
func (fb *Firebase) RemoveAuth() {
	fb.auth = ""
}
