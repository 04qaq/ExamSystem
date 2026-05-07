package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	RealName string `json:"real_name" binding:"max=50"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterResponse struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Role     int8   `json:"role"`
	RealName string `json:"real_name"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Role     int8   `json:"role"`
	RealName string `json:"real_name"`
}
