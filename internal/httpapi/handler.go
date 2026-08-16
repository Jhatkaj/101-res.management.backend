package httpapi

import (
	"encoding/json"
	"github.com/jhatkaz/restaurant-console/internal/model"
	"github.com/jhatkaz/restaurant-console/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{ app *service.Restaurant }

func NewHandler(app *service.Restaurant) *Handler { return &Handler{app: app} }
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (h *Handler) setupStatus(w http.ResponseWriter, r *http.Request) {
	complete, e := h.app.SetupComplete(r.Context())
	respond(w, map[string]bool{"configured": complete}, e)
}
func (h *Handler) setupRestaurant(w http.ResponseWriter, r *http.Request) {
	var body model.SetupRequest
	if !decode(w, r, &body) {
		return
	}
	restaurant, owner, e := h.app.Setup(r.Context(), body)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, e.Error())
		return
	}
	auth, e := h.app.CreateSession(owner)
	if e != nil {
		errorJSON(w, http.StatusInternalServerError, "restaurant created but owner session could not be started")
		return
	}
	respondStatus(w, map[string]any{"restaurant": restaurant, "owner": owner, "auth": auth}, nil, http.StatusCreated)
}
func (h *Handler) branches(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "owner", "hr_vp", "hr_manager") {
		return
	}
	v, e := h.app.Restaurants(r.Context())
	respond(w, v, e)
}
func (h *Handler) createBranch(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "owner") {
		return
	}
	var body model.BranchRequest
	if !decode(w, r, &body) {
		return
	}
	branch, e := h.app.CreateBranch(r.Context(), body)
	respondStatus(w, branch, e, http.StatusCreated)
}
func (h *Handler) updateBranch(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "owner") {
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, "invalid branch id")
		return
	}
	var body model.BranchRequest
	if !decode(w, r, &body) {
		return
	}
	branch, e := h.app.UpdateBranch(r.Context(), id, body)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, e.Error())
		return
	}
	respond(w, branch, nil)
}
func (h *Handler) bootstrap(w http.ResponseWriter, r *http.Request) {
	menu, e := h.app.Menu(r.Context())
	if e != nil {
		respond(w, nil, e)
		return
	}
	specials, e := h.app.Specials(r.Context())
	if e != nil {
		respond(w, nil, e)
		return
	}
	result := model.Bootstrap{Menu: menu, Specials: specials}
	if user, ok := h.user(r); ok && canManageOrders(user.Role) {
		result.Orders, e = h.app.Orders(r.Context())
		if e == nil {
			result.Stock, e = h.app.Stock(r.Context())
		}
	}
	respond(w, result, e)
}
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body model.RegisterRequest
	if !decode(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Phone) == "" {
		errorJSON(w, http.StatusBadRequest, "phone is required")
		return
	}
	user, e := h.app.CreateUser(r.Context(), body.Username, body.Password, body.Name, "customer", body.Phone)
	respondStatus(w, user, e, http.StatusCreated)
}
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body model.AuthRequest
	if !decode(w, r, &body) {
		return
	}
	v, e := h.app.Authenticate(r.Context(), body.Username, body.Password)
	if e != nil {
		errorJSON(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	respond(w, v, e)
}
func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := h.user(r)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var body model.ChangePasswordRequest
	if !decode(w, r, &body) {
		return
	}
	if e := h.app.ChangePassword(r.Context(), user, body.CurrentPassword, body.NewPassword); e != nil {
		errorJSON(w, http.StatusBadRequest, "current password is incorrect or new password is too short")
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
func (h *Handler) users(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.user(r)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if !hasRole(actor.Role, "owner", "hr_vp", "hr_manager", "branch_manager", "admin", "staff_manager") {
		errorJSON(w, http.StatusForbidden, "you do not have permission for staff management")
		return
	}
	v, e := h.app.UsersFor(r.Context(), actor)
	respond(w, v, e)
}
func (h *Handler) deleteStaff(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.user(r)
	if !ok || !hasRole(actor.Role, "owner", "hr_vp", "hr_manager", "branch_manager", "admin", "staff_manager") {
		errorJSON(w, http.StatusForbidden, "you do not have permission for staff management")
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	target, e := h.app.UserByID(r.Context(), id)
	allowed := actor.Role == "owner" && target.Role == "hr_vp"
	if !allowed {
		allowed = canManageStaff(actor, target)
	}
	if e != nil || !allowed {
		errorJSON(w, http.StatusForbidden, "you cannot delete this account")
		return
	}
	if e = h.app.DeleteUser(r.Context(), id); e != nil {
		errorJSON(w, http.StatusBadRequest, e.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) createStaff(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.user(r)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if !hasRole(actor.Role, "owner", "hr_vp", "hr_manager", "branch_manager", "admin", "staff_manager") {
		errorJSON(w, http.StatusForbidden, "you do not have permission to create this account")
		return
	}
	var body model.CreateStaffRequest
	if !decode(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Phone) == "" || strings.TrimSpace(body.Aadhar) == "" || strings.TrimSpace(body.Address) == "" || strings.TrimSpace(body.Branch) == "" {
		errorJSON(w, http.StatusBadRequest, "phone, aadhar, address, and branch are required")
		return
	}
	if body.Role == "other" && strings.TrimSpace(body.StaffType) == "" {
		errorJSON(w, http.StatusBadRequest, "custom staff type is required for Other")
		return
	}
	if !canCreateRole(actor, body.Role) || !centralStaffRole(actor.Role) && body.Branch != actor.Branch {
		errorJSON(w, http.StatusForbidden, "you cannot create this role or branch")
		return
	}
	parentID := body.ParentID
	if parentID == 0 {
		parentID = actor.ID
	}
	parent, parentErr := h.app.UserByID(r.Context(), parentID)
	if parentErr != nil || parent.Role == "customer" || parent.ID == 0 {
		errorJSON(w, http.StatusBadRequest, "invalid reports-to account")
		return
	}
	user, e := h.app.CreateUser(r.Context(), body.Username, body.Password, body.Name, body.Role, body.Phone, body.Aadhar, body.Address, body.ImageURL, body.Branch, strconv.FormatInt(parentID, 10), body.StaffType)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user details or role")
		return
	}
	respondStatus(w, user, e, http.StatusCreated)
}
func (h *Handler) updateUserStatus(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.user(r)
	if !ok || !hasRole(actor.Role, "hr_vp", "hr_manager", "branch_manager", "admin", "staff_manager") {
		errorJSON(w, http.StatusForbidden, "you do not have permission for staff management")
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	target, e := h.app.UserByID(r.Context(), id)
	if e != nil || !canManageStaff(actor, target) {
		errorJSON(w, http.StatusForbidden, "you cannot change this account")
		return
	}
	var body struct {
		Active bool `json:"active"`
	}
	if !decode(w, r, &body) {
		return
	}
	user, e := h.app.SetUserActive(r.Context(), id, body.Active)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, e.Error())
		return
	}
	respond(w, user, e)
}
func (h *Handler) updateStaff(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.user(r)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if !hasRole(actor.Role, "hr_vp", "hr_manager") {
		errorJSON(w, http.StatusForbidden, "only HR users can change staff assignments")
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	target, e := h.app.UserByID(r.Context(), id)
	if e != nil || !canManageStaff(actor, target) {
		errorJSON(w, http.StatusForbidden, "you cannot change this account")
		return
	}
	var body struct {
		Name      string `json:"name"`
		StaffType string `json:"staffType"`
		Phone     string `json:"phone"`
		Aadhar    string `json:"aadhar"`
		Address   string `json:"address"`
		ImageURL  string `json:"imageUrl"`
		Branch    string `json:"branch"`
		ParentID  int64  `json:"parentId"`
	}
	if !decode(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Name) == "" || strings.TrimSpace(body.Branch) == "" || strings.TrimSpace(body.Phone) == "" || strings.TrimSpace(body.Aadhar) == "" || strings.TrimSpace(body.Address) == "" || body.ParentID == id {
		errorJSON(w, http.StatusBadRequest, "name, branch, and a valid parent account are required")
		return
	}
	if target.Role == "other" && strings.TrimSpace(body.StaffType) == "" {
		errorJSON(w, http.StatusBadRequest, "custom staff type is required for Other")
		return
	}
	if body.ParentID != 0 {
		parent, parentErr := h.app.UserByID(r.Context(), body.ParentID)
		if parentErr != nil || parent.Role == "customer" || parent.ID == target.ID {
			errorJSON(w, http.StatusBadRequest, "invalid parent account")
			return
		}
	}
	updated, e := h.app.UpdateUserAssignment(r.Context(), id, strings.TrimSpace(body.Name), strings.TrimSpace(body.Branch), strings.TrimSpace(body.StaffType), strings.TrimSpace(body.Phone), strings.TrimSpace(body.Aadhar), strings.TrimSpace(body.Address), strings.TrimSpace(body.ImageURL), body.ParentID)
	respond(w, updated, e)
}
func (h *Handler) salary(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "admin", "owner") {
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	v, e := h.app.SalaryPayments(r.Context(), id)
	respond(w, v, e)
}
func (h *Handler) paySalary(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "admin", "owner") {
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var body struct {
		Month  string  `json:"month"`
		Amount float64 `json:"amount"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Month == "" || body.Amount < 0 {
		errorJSON(w, http.StatusBadRequest, "month and amount are required")
		return
	}
	v, e := h.app.PaySalary(r.Context(), model.SalaryPayment{UserID: id, Month: body.Month, Amount: body.Amount})
	respondStatus(w, v, e, http.StatusCreated)
}
func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "admin", "owner") {
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var body struct {
		NewPassword string `json:"newPassword"`
	}
	if !decode(w, r, &body) {
		return
	}
	if e = h.app.ResetPassword(r.Context(), id, body.NewPassword); e != nil {
		errorJSON(w, http.StatusBadRequest, "new password must be at least 4 characters")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) menu(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "order_manager", "special_admin", "owner") {
		return
	}
	v, e := h.app.Menu(r.Context())
	respond(w, v, e)
}
func (h *Handler) specials(w http.ResponseWriter, r *http.Request) {
	v, e := h.app.Specials(r.Context())
	respond(w, v, e)
}
func (h *Handler) stock(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "stock_manager", "special_admin", "owner") {
		return
	}
	v, e := h.app.Stock(r.Context())
	respond(w, v, e)
}
func (h *Handler) orders(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "order_manager", "special_admin", "owner") {
		return
	}
	v, e := h.app.Orders(r.Context())
	respond(w, v, e)
}
func (h *Handler) myOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := h.user(r)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if user.Role != "customer" {
		errorJSON(w, http.StatusForbidden, "only customers can access personal order history")
		return
	}
	v, e := h.app.OrdersByCustomer(r.Context(), user.Name)
	respond(w, v, e)
}
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := h.user(r)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if user.Role == "admin" {
		errorJSON(w, http.StatusForbidden, "admin accounts only manage staff")
		return
	}
	var v model.Order
	if !decode(w, r, &v) {
		return
	}
	if v.ID == "" || len(v.Items) == 0 {
		errorJSON(w, http.StatusBadRequest, "id and items are required")
		return
	}
	v.Customer = user.Name
	saved, e := h.app.CreateOrder(r.Context(), v)
	respondStatus(w, saved, e, http.StatusCreated)
}
func (h *Handler) updateOrder(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "order_manager", "special_admin", "owner") {
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Status == "" {
		errorJSON(w, http.StatusBadRequest, "status is required")
		return
	}
	v, e := h.app.UpdateOrder(r.Context(), r.PathValue("id"), body.Status)
	respond(w, v, e)
}
func (h *Handler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "order_manager", "special_admin", "owner") {
		return
	}
	e := h.app.DeleteOrder(r.Context(), r.PathValue("id"))
	if e != nil {
		errorJSON(w, http.StatusNotFound, "order not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) createMenu(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "order_manager", "special_admin") {
		return
	}
	var v model.MenuItem
	if !decode(w, r, &v) {
		return
	}
	saved, e := h.app.CreateMenu(r.Context(), v)
	respondStatus(w, saved, e, http.StatusCreated)
}
func (h *Handler) deleteMenu(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "order_manager", "special_admin") {
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, 400, "invalid id")
		return
	}
	if e = h.app.DeleteMenu(r.Context(), id); e != nil {
		errorJSON(w, 404, "menu item not found")
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) createSpecial(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "order_manager", "special_admin") {
		return
	}
	var v model.Special
	if !decode(w, r, &v) {
		return
	}
	saved, e := h.app.CreateSpecial(r.Context(), v)
	respondStatus(w, saved, e, 201)
}
func (h *Handler) deleteSpecial(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "order_manager", "special_admin") {
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, 400, "invalid id")
		return
	}
	if e = h.app.DeleteSpecial(r.Context(), id); e != nil {
		errorJSON(w, 404, "special not found")
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) createStock(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "stock_manager", "special_admin") {
		return
	}
	var v model.StockItem
	if !decode(w, r, &v) {
		return
	}
	saved, e := h.app.CreateStock(r.Context(), v)
	respondStatus(w, saved, e, 201)
}
func (h *Handler) updateStock(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "stock_manager", "special_admin") {
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, 400, "invalid id")
		return
	}
	var v model.StockItem
	if !decode(w, r, &v) {
		return
	}
	saved, e := h.app.UpdateStock(r.Context(), id, v)
	respond(w, saved, e)
}
func (h *Handler) deleteStock(w http.ResponseWriter, r *http.Request) {
	if !h.requireRole(w, r, "stock_manager", "special_admin") {
		return
	}
	id, e := idParam(r)
	if e != nil {
		errorJSON(w, 400, "invalid id")
		return
	}
	if e = h.app.DeleteStock(r.Context(), id); e != nil {
		errorJSON(w, 404, "stock item not found")
		return
	}
	w.WriteHeader(204)
}
func idParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	if e := json.NewDecoder(r.Body).Decode(v); e != nil {
		errorJSON(w, 400, "invalid JSON body")
		return false
	}
	return true
}
func respond(w http.ResponseWriter, v any, e error) {
	if e != nil {
		errorJSON(w, 500, "internal server error")
		return
	}
	writeJSON(w, 200, v)
}
func respondStatus(w http.ResponseWriter, v any, e error, status int) {
	if e != nil {
		errorJSON(w, 500, "internal server error")
		return
	}
	writeJSON(w, status, v)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func errorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func (h *Handler) user(r *http.Request) (model.User, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return model.User{}, false
	}
	return h.app.UserFromToken(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
}
func canManageOrders(role string) bool {
	return role == "order_manager" || role == "special_admin" || role == "owner"
}
func hasRole(role string, allowed ...string) bool {
	for _, candidate := range allowed {
		if role == candidate {
			return true
		}
	}
	return false
}
func canCreateRole(actor model.User, targetRole string) bool {
	switch actor.Role {
	case "owner":
		return targetRole == "hr_vp"
	case "hr_vp":
		return targetRole == "hr_manager"
	case "hr_manager":
		return targetRole == "branch_manager" || targetRole == "admin" || targetRole == "staff_manager" || targetRole == "order_manager" || targetRole == "stock_manager" || targetRole == "other"
	case "branch_manager":
		return targetRole == "admin" || targetRole == "staff_manager"
	case "admin":
		return targetRole == "staff_manager" || targetRole == "order_manager" || targetRole == "stock_manager"
	case "staff_manager":
		return targetRole == "order_manager" || targetRole == "stock_manager"
	default:
		return false
	}
}
func canManageStaff(actor, target model.User) bool {
	if target.Role == "owner" || target.Role == "customer" || target.ID == actor.ID || !centralStaffRole(actor.Role) && actor.Branch != target.Branch {
		return false
	}
	switch actor.Role {
	case "owner":
		return false
	case "hr_vp":
		return target.Role == "hr_manager"
	case "hr_manager":
		return target.Role == "branch_manager" || target.Role == "admin" || target.Role == "staff_manager" || target.Role == "order_manager" || target.Role == "stock_manager" || target.Role == "other"
	case "branch_manager":
		return target.Role == "admin" || target.Role == "staff_manager" || target.Role == "order_manager" || target.Role == "stock_manager"
	case "admin":
		return target.Role == "staff_manager" || target.Role == "order_manager" || target.Role == "stock_manager"
	case "staff_manager":
		return target.Role == "order_manager" || target.Role == "stock_manager"
	default:
		return false
	}
}
func centralStaffRole(role string) bool {
	return role == "owner" || role == "hr_vp" || role == "hr_manager"
}
func (h *Handler) requireRole(w http.ResponseWriter, r *http.Request, roles ...string) bool {
	user, ok := h.user(r)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	for _, role := range roles {
		if user.Role == role {
			return true
		}
	}
	errorJSON(w, http.StatusForbidden, "you do not have permission for this resource")
	return false
}
