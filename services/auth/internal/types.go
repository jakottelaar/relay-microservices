package internal

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"` // Expiration time in seconds
	Account      *Account `json:"account"`
}

type SessionMetadata struct {
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
}