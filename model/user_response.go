package model

// UserAdminResponse - Estructura para respuesta de admin (sin contraseña)
type UserAdminResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	RoleID    int    `json:"role_id"`
	RoleName  string `json:"role_name"`
	CreatedAt string `json:"created_at"`
}

// UserPaginatedResponse - Respuesta paginada para admin
type UserPaginatedResponse struct {
	Data       []UserAdminResponse `json:"data"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
	HasNext    bool                `json:"has_next"`
	HasPrev    bool                `json:"has_prev"`
}
