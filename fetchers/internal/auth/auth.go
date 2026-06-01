package auth

// BasicAuth carries HTTP Basic Authentication credentials.
type BasicAuth struct {
	Username string
	Password string
}

// Clone returns a detached copy of auth credentials.
func Clone(value *BasicAuth) *BasicAuth {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
