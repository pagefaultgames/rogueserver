package api

// error

type ErrorResponse struct{
	Message string `json:"message"`
}

// auth

type GenericAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GenericAuthResponse struct {
	Token string `json:"token"`
}
