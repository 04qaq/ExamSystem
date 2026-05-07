package dto

// 用户管理
type UserListItem struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Role      int8   `json:"role"`
	RealName  string `json:"real_name"`
	Status    int8   `json:"status"`
	CreatedAt string `json:"created_at"`
}

type UserListResponse struct {
	Total int64           `json:"total"`
	Items []UserListItem  `json:"items"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     int8   `json:"role" binding:"required"`
	RealName string `json:"real_name"`
}

type UpdateUserRequest struct {
	RealName string `json:"real_name"`
	Role     *int8  `json:"role"`
	Status   *int8  `json:"status"`
	Password string `json:"password"`
}

// 操作日志
type OperationLogItem struct {
	ID        uint64 `json:"id"`
	UserID    uint64 `json:"user_id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Detail    string `json:"detail"`
	IP        string `json:"ip"`
	CreatedAt string `json:"created_at"`
}

type LogListResponse struct {
	Total int64              `json:"total"`
	Items []OperationLogItem `json:"items"`
}
