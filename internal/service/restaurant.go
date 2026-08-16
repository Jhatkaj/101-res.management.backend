package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/jhatkaz/restaurant-console/internal/model"
	"github.com/jhatkaz/restaurant-console/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"sync"
)

type Restaurant struct {
	repo       *repository.Restaurant
	sessions   map[string]model.User
	sessionsMu sync.RWMutex
}

func NewRestaurant(repo *repository.Restaurant) *Restaurant {
	return &Restaurant{repo: repo, sessions: make(map[string]model.User)}
}
func (s *Restaurant) SetupComplete(ctx context.Context) (bool, error) {
	return s.repo.SetupComplete(ctx)
}
func (s *Restaurant) CreateSession(user model.User) (model.AuthResponse, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return model.AuthResponse{}, err
	}
	token := hex.EncodeToString(raw)
	user.PasswordHash = ""
	s.sessionsMu.Lock()
	s.sessions[token] = user
	s.sessionsMu.Unlock()
	return model.AuthResponse{Token: token, User: user}, nil
}
func (s *Restaurant) Setup(ctx context.Context, request model.SetupRequest) (model.Restaurant, model.User, error) {
	for _, value := range []string{request.RestaurantName, request.OwnerName, request.Username, request.Password} {
		if strings.TrimSpace(value) == "" {
			return model.Restaurant{}, model.User{}, errors.New("all required setup fields must be provided")
		}
	}
	if len(request.Password) < 4 {
		return model.Restaurant{}, model.User{}, errors.New("password must be at least 4 characters")
	}
	complete, err := s.SetupComplete(ctx)
	if err != nil {
		return model.Restaurant{}, model.User{}, err
	}
	if complete {
		return model.Restaurant{}, model.User{}, errors.New("restaurant is already configured")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.Restaurant{}, model.User{}, err
	}
	return s.repo.CreateRestaurant(ctx, model.Restaurant{Name: strings.TrimSpace(request.RestaurantName), Branch: "Main Branch", Address: strings.TrimSpace(request.Address), Phone: strings.TrimSpace(request.Phone)}, model.User{Username: strings.TrimSpace(request.Username), Name: strings.TrimSpace(request.OwnerName), PasswordHash: string(hash), Role: "owner"})
}

func (s *Restaurant) CreateBranch(ctx context.Context, request model.BranchRequest) (model.Restaurant, error) {
	name := strings.TrimSpace(request.Name)
	branch := strings.TrimSpace(request.Branch)
	if name == "" || branch == "" {
		return model.Restaurant{}, errors.New("restaurant name and branch are required")
	}
	return s.repo.CreateBranch(ctx, model.Restaurant{Name: name, Branch: branch, Address: strings.TrimSpace(request.Address), Phone: strings.TrimSpace(request.Phone)})
}
func (s *Restaurant) UpdateBranch(ctx context.Context, id int64, request model.BranchRequest) (model.Restaurant, error) {
	if strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.Branch) == "" {
		return model.Restaurant{}, errors.New("restaurant name and branch are required")
	}
	return s.repo.UpdateBranch(ctx, id, model.Restaurant{Name: strings.TrimSpace(request.Name), Branch: strings.TrimSpace(request.Branch), Address: strings.TrimSpace(request.Address), Phone: strings.TrimSpace(request.Phone)})
}
func (s *Restaurant) Restaurants(ctx context.Context) ([]model.Restaurant, error) {
	return s.repo.Restaurants(ctx)
}

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidRole = errors.New("invalid role")

func validRole(role string) bool {
	switch role {
	case "order_manager", "stock_manager", "staff_manager", "admin", "branch_manager", "special_admin", "hr_vp", "hr_manager", "other", "owner":
		return true
	}
	return false
}
func (s *Restaurant) CreateUser(ctx context.Context, username, password, name, role string, details ...string) (model.User, error) {
	username = strings.TrimSpace(username)
	name = strings.TrimSpace(name)
	if username == "" || name == "" || len(password) < 4 || !validRole(role) {
		return model.User{}, ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	user := model.User{Username: username, Name: name, PasswordHash: string(hash), Role: role}
	if len(details) > 0 {
		user.Phone = details[0]
	}
	if len(details) > 1 {
		user.Aadhar = details[1]
	}
	if len(details) > 2 {
		user.Address = details[2]
	}
	if len(details) > 3 {
		user.ImageURL = details[3]
	}
	if len(details) > 4 {
		user.Branch = strings.TrimSpace(details[4])
	}
	if user.Branch == "" {
		user.Branch = "Main Branch"
	}
	if len(details) > 5 && strings.TrimSpace(details[5]) != "" {
		user.ParentID, err = strconv.ParseInt(strings.TrimSpace(details[5]), 10, 64)
		if err != nil {
			return model.User{}, errors.New("invalid parent account")
		}
	}
	if len(details) > 6 {
		user.StaffType = strings.TrimSpace(details[6])
	}
	if role == "owner" {
		existing, err := s.repo.Users(ctx)
		if err != nil {
			return model.User{}, err
		}
		for _, account := range existing {
			if account.Role == "owner" {
				return model.User{}, errors.New("only one owner account is allowed")
			}
		}
	}
	return s.repo.CreateUser(ctx, user)
}
func (s *Restaurant) Authenticate(ctx context.Context, username, password string) (model.AuthResponse, error) {
	user, err := s.repo.User(ctx, strings.TrimSpace(username))
	if err != nil {
		return model.AuthResponse{}, ErrInvalidCredentials
	}
	if user.Role == "customer" || !user.Active {
		return model.AuthResponse{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.AuthResponse{}, ErrInvalidCredentials
	}
	return s.CreateSession(user)
}
func (s *Restaurant) ChangePassword(ctx context.Context, user model.User, currentPassword, newPassword string) error {
	stored, err := s.repo.User(ctx, user.Username)
	if err != nil || len(newPassword) < 4 || bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte(currentPassword)) != nil {
		return ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, user.ID, string(hash))
}
func (s *Restaurant) ResetPassword(ctx context.Context, id int64, newPassword string) error {
	if len(newPassword) < 4 {
		return ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, id, string(hash))
}
func (s *Restaurant) UserFromToken(token string) (model.User, bool) {
	s.sessionsMu.RLock()
	user, ok := s.sessions[token]
	s.sessionsMu.RUnlock()
	return user, ok && user.Active
}
func (s *Restaurant) Users(ctx context.Context) ([]model.User, error) { return s.repo.Users(ctx) }
func (s *Restaurant) UsersFor(ctx context.Context, actor model.User) ([]model.User, error) {
	users, err := s.repo.Users(ctx)
	if err != nil {
		return users, err
	}
	if actor.Role == "owner" || actor.Role == "hr_vp" || actor.Role == "hr_manager" {
		result := make([]model.User, 0, len(users))
		for _, user := range users {
			if user.Role != "customer" {
				result = append(result, user)
			}
		}
		return result, nil
	}
	result := make([]model.User, 0)
	for _, user := range users {
		if user.Branch != actor.Branch {
			continue
		}
		switch actor.Role {
		case "branch_manager":
			if user.Role != "owner" {
				result = append(result, user)
			}
		case "admin":
			if user.Role != "owner" && user.Role != "branch_manager" && user.Role != "admin" {
				result = append(result, user)
			}
		case "staff_manager":
			if user.Role == "order_manager" || user.Role == "stock_manager" {
				result = append(result, user)
			}
		}
	}
	return result, nil
}
func (s *Restaurant) UserByID(ctx context.Context, id int64) (model.User, error) {
	return s.repo.UserByID(ctx, id)
}
func (s *Restaurant) UpdateUser(ctx context.Context, id int64, name, branch string) (model.User, error) {
	return s.repo.UpdateUser(ctx, id, name, branch)
}
func (s *Restaurant) UpdateUserAssignment(ctx context.Context, id int64, name, branch, staffType, phone, aadhar, address, imageURL string, parentID int64) (model.User, error) {
	return s.repo.UpdateUserAssignment(ctx, id, name, branch, staffType, phone, aadhar, address, imageURL, parentID)
}
func (s *Restaurant) SetUserActive(ctx context.Context, id int64, active bool) (model.User, error) {
	existing, err := s.repo.UserByID(ctx, id)
	if err != nil || existing.Role == "owner" {
		return existing, errors.New("owner accounts must remain active")
	}
	user, err := s.repo.SetUserActive(ctx, id, active)
	if err != nil {
		return user, err
	}
	s.sessionsMu.Lock()
	for token, session := range s.sessions {
		if session.ID == id {
			session.Active = active
			s.sessions[token] = session
		}
	}
	s.sessionsMu.Unlock()
	return user, nil
}
func (s *Restaurant) DeleteUser(ctx context.Context, id int64) error {
	existing, err := s.repo.UserByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.Role == "customer" || existing.Role == "owner" {
		return errors.New("only staff accounts can be deleted")
	}
	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return err
	}
	s.sessionsMu.Lock()
	for token, session := range s.sessions {
		if session.ID == id {
			delete(s.sessions, token)
		}
	}
	s.sessionsMu.Unlock()
	return nil
}
func (s *Restaurant) PaySalary(ctx context.Context, payment model.SalaryPayment) (model.SalaryPayment, error) {
	return s.repo.PaySalary(ctx, payment)
}
func (s *Restaurant) SalaryPayments(ctx context.Context, id int64) ([]model.SalaryPayment, error) {
	return s.repo.SalaryPayments(ctx, id)
}
func (s *Restaurant) Bootstrap(ctx context.Context) (model.Bootstrap, error) {
	menu, e := s.repo.Menu(ctx)
	if e != nil {
		return model.Bootstrap{}, e
	}
	specials, e := s.repo.Specials(ctx)
	if e != nil {
		return model.Bootstrap{}, e
	}
	orders, e := s.repo.Orders(ctx)
	if e != nil {
		return model.Bootstrap{}, e
	}
	stock, e := s.repo.Stock(ctx)
	if e != nil {
		return model.Bootstrap{}, e
	}
	return model.Bootstrap{Menu: menu, Specials: specials, Orders: orders, Stock: stock}, nil
}
func (s *Restaurant) Menu(ctx context.Context) ([]model.MenuItem, error) { return s.repo.Menu(ctx) }
func (s *Restaurant) Specials(ctx context.Context) ([]model.Special, error) {
	return s.repo.Specials(ctx)
}
func (s *Restaurant) Stock(ctx context.Context) ([]model.StockItem, error) { return s.repo.Stock(ctx) }
func (s *Restaurant) Orders(ctx context.Context) ([]model.Order, error)    { return s.repo.Orders(ctx) }
func (s *Restaurant) OrdersByCustomer(ctx context.Context, customer string) ([]model.Order, error) {
	return s.repo.OrdersByCustomer(ctx, customer)
}
func (s *Restaurant) CreateOrder(ctx context.Context, o model.Order) (model.Order, error) {
	return s.repo.CreateOrder(ctx, o)
}
func (s *Restaurant) UpdateOrder(ctx context.Context, id, status string) (model.Order, error) {
	return s.repo.UpdateOrder(ctx, id, status)
}
func (s *Restaurant) DeleteOrder(ctx context.Context, id string) error {
	return s.repo.DeleteOrder(ctx, id)
}
func (s *Restaurant) CreateMenu(ctx context.Context, i model.MenuItem) (model.MenuItem, error) {
	return s.repo.CreateMenu(ctx, i)
}
func (s *Restaurant) DeleteMenu(ctx context.Context, id int64) error {
	return s.repo.DeleteMenu(ctx, id)
}
func (s *Restaurant) CreateSpecial(ctx context.Context, i model.Special) (model.Special, error) {
	return s.repo.CreateSpecial(ctx, i)
}
func (s *Restaurant) DeleteSpecial(ctx context.Context, id int64) error {
	return s.repo.DeleteSpecial(ctx, id)
}
func (s *Restaurant) CreateStock(ctx context.Context, i model.StockItem) (model.StockItem, error) {
	return s.repo.CreateStock(ctx, i)
}
func (s *Restaurant) UpdateStock(ctx context.Context, id int64, i model.StockItem) (model.StockItem, error) {
	return s.repo.UpdateStock(ctx, id, i)
}
func (s *Restaurant) DeleteStock(ctx context.Context, id int64) error {
	return s.repo.DeleteStock(ctx, id)
}
