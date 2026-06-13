package dto

import "utfpr.edu.br/carshop-api/src/domain"

// UserRequest is the POST/PUT body for /api/v1/user.
// The C# contract uses PascalCase role names ("Admin"/"Vendor") on the wire;
// internally we normalize to lowercase.
type UserRequest struct {
	Username                  string `json:"username" binding:"required"`
	Password                  string `json:"password" binding:"required"`
	Email                     string `json:"email" binding:"required"`
	ComissionPerSaleInPercent *int16 `json:"comissionPerSaleInPercent"`
	Role                      string `json:"role" binding:"required,oneof=Admin Vendor"`
}

// UserResponse intentionally omits Password so we never leak the bcrypt
// hash (or any prior plaintext that snuck through). This is the documented
// contract drift from the C# version, where User.Password was echoed in
// GET responses.
type UserResponse struct {
	ID                        int64  `json:"id"`
	Username                  string `json:"username"`
	Email                     string `json:"email"`
	ComissionPerSaleInPercent *int16 `json:"comissionPerSaleInPercent"`
	Role                      string `json:"role"`
}

func (r UserRequest) ToDomain() domain.User {
	return domain.User{
		Username:                  r.Username,
		Password:                  r.Password,
		Email:                     r.Email,
		ComissionPerSaleInPercent: r.ComissionPerSaleInPercent,
		Role:                      RoleFromDTO(r.Role),
	}
}

func UserToResponse(u domain.User) UserResponse {
	return UserResponse{
		ID:                        u.ID,
		Username:                  u.Username,
		Email:                     u.Email,
		ComissionPerSaleInPercent: u.ComissionPerSaleInPercent,
		Role:                      RoleToDTO(u.Role),
	}
}

func RoleFromDTO(s string) domain.UserRole {
	switch s {
	case "Admin":
		return domain.RoleAdmin
	case "Vendor":
		return domain.RoleVendor
	}
	return ""
}

func RoleToDTO(r domain.UserRole) string {
	switch r {
	case domain.RoleAdmin:
		return "Admin"
	case domain.RoleVendor:
		return "Vendor"
	}
	return ""
}
