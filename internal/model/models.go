package model

import "time"

type MenuItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Emoji       string  `json:"emoji"`
}
type Special struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Emoji       string  `json:"emoji"`
}
type StockItem struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	Icon     string  `json:"icon"`
}
type OrderItem struct {
	ID    int64   `json:"id,omitempty"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Qty   int     `json:"qty"`
	Emoji string  `json:"emoji,omitempty"`
}
type Order struct {
	ID        string      `json:"id"`
	Customer  string      `json:"customer"`
	Date      string      `json:"date"`
	CreatedAt time.Time   `json:"createdAt"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	Status    string      `json:"status"`
}
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	StaffType    string    `json:"staffType,omitempty"`
	Branch       string    `json:"branch"`
	ParentID     int64     `json:"parentId,omitempty"`
	Phone        string    `json:"phone"`
	Aadhar       string    `json:"aadhar"`
	Address      string    `json:"address"`
	ImageURL     string    `json:"imageUrl"`
	Active       bool      `json:"active"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}
type Restaurant struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Branch    string    `json:"branch"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"createdAt"`
}
type SetupRequest struct {
	RestaurantName string `json:"restaurantName"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	OwnerName      string `json:"ownerName"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}
type BranchRequest struct {
	Name    string `json:"name"`
	Branch  string `json:"branch"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}
type CreateStaffRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	StaffType string `json:"staffType"`
	Phone     string `json:"phone"`
	Aadhar    string `json:"aadhar"`
	Address   string `json:"address"`
	ImageURL  string `json:"imageUrl"`
	Branch    string `json:"branch"`
	ParentID  int64  `json:"parentId,omitempty"`
}
type SalaryPayment struct {
	ID     int64     `json:"id"`
	UserID int64     `json:"userId"`
	Month  string    `json:"month"`
	Amount float64   `json:"amount"`
	PaidAt time.Time `json:"paidAt"`
}
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
type Bootstrap struct {
	Menu     []MenuItem  `json:"menu"`
	Specials []Special   `json:"specials"`
	Orders   []Order     `json:"orders,omitempty"`
	Stock    []StockItem `json:"stock,omitempty"`
}
