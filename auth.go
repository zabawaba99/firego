package firego

// SetAuth sets the custom Firebase token used to authenticate to Firebase
func (fb *Firebase) Auth(token string) {
	fb.auth = token
}

// RemoveAuth removes the current token being used to authenticate to Firebase
func (fb *Firebase) Unauth() {
	fb.auth = ""
}
