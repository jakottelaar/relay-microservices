package internal

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Username string `json:"username" binding:"required,min=3,max=32"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
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

type SessionResponse struct {
	SessionID          int64  `json:"session_id"`
	ApproxLastTimeUsed string `json:"approx_last_time_used"`
	Client             struct {
		UserAgent string `json:"user_agent"`
	} `json:"client"`
}