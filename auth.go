package firego

// Auth sets the custom Firebase token used to authenticate to Firebase
func (fb *Firebase) Auth(token string) {
	fb.params.Set(authParam, token)
}

// Unauth removes the current token being used to authenticate to Firebase
func (fb *Firebase) Unauth() {
	fb.params.Del(authParam)
}
