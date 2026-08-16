package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jhatkaz/restaurant-console/internal/model"
)

type Restaurant struct{ db *pgxpool.Pool }

func nullableID(id int64) any {
	if id == 0 {
		return nil
	}
	return id
}

func New(db *pgxpool.Pool) *Restaurant { return &Restaurant{db: db} }

func (r *Restaurant) SetupComplete(ctx context.Context) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM restaurants)`).Scan(&exists)
	return exists, err
}
func (r *Restaurant) CreateRestaurant(ctx context.Context, restaurant model.Restaurant, owner model.User) (model.Restaurant, model.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return restaurant, owner, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `INSERT INTO restaurants(name,branch,address,phone) VALUES($1,$2,$3,$4) RETURNING id,created_at`, restaurant.Name, restaurant.Branch, restaurant.Address, restaurant.Phone).Scan(&restaurant.ID, &restaurant.CreatedAt)
	if err != nil {
		return restaurant, owner, err
	}
	err = tx.QueryRow(ctx, `UPDATE users SET username=$1,name=$2,password_hash=$3,branch='Main Branch',active=TRUE WHERE role='owner' RETURNING id,active,created_at`, owner.Username, owner.Name, owner.PasswordHash).Scan(&owner.ID, &owner.Active, &owner.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `INSERT INTO users(username,name,password_hash,role,branch,active) VALUES($1,$2,$3,'owner','Main Branch',TRUE) RETURNING id,active,created_at`, owner.Username, owner.Name, owner.PasswordHash).Scan(&owner.ID, &owner.Active, &owner.CreatedAt)
	}
	if err != nil {
		return restaurant, owner, err
	}
	if err = tx.Commit(ctx); err != nil {
		return restaurant, owner, err
	}
	owner.Branch = "Main Branch"
	return restaurant, owner, nil
}

func (r *Restaurant) CreateBranch(ctx context.Context, branch model.Restaurant) (model.Restaurant, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO restaurants(name,branch,address,phone) VALUES($1,$2,$3,$4) RETURNING id,created_at`, branch.Name, branch.Branch, branch.Address, branch.Phone).Scan(&branch.ID, &branch.CreatedAt)
	return branch, err
}
func (r *Restaurant) UpdateBranch(ctx context.Context, id int64, branch model.Restaurant) (model.Restaurant, error) {
	err := r.db.QueryRow(ctx, `UPDATE restaurants SET name=$2,branch=$3,address=$4,phone=$5 WHERE id=$1 RETURNING id,name,branch,address,phone,created_at`, id, branch.Name, branch.Branch, branch.Address, branch.Phone).Scan(&branch.ID, &branch.Name, &branch.Branch, &branch.Address, &branch.Phone, &branch.CreatedAt)
	return branch, err
}

func (r *Restaurant) Restaurants(ctx context.Context) ([]model.Restaurant, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,branch,address,phone,created_at FROM restaurants ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Restaurant{}
	for rows.Next() {
		var branch model.Restaurant
		if err := rows.Scan(&branch.ID, &branch.Name, &branch.Branch, &branch.Address, &branch.Phone, &branch.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, branch)
	}
	return result, rows.Err()
}

func (r *Restaurant) User(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, `SELECT id,username,name,password_hash,role,staff_type,phone,aadhar,address,image_url,branch,COALESCE(parent_id,0),active,created_at FROM users WHERE lower(username)=lower($1)`, username).Scan(&user.ID, &user.Username, &user.Name, &user.PasswordHash, &user.Role, &user.StaffType, &user.Phone, &user.Aadhar, &user.Address, &user.ImageURL, &user.Branch, &user.ParentID, &user.Active, &user.CreatedAt)
	return user, err
}
func (r *Restaurant) UserByID(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, `SELECT id,username,name,password_hash,role,staff_type,phone,aadhar,address,image_url,branch,COALESCE(parent_id,0),active,created_at FROM users WHERE id=$1`, id).Scan(&user.ID, &user.Username, &user.Name, &user.PasswordHash, &user.Role, &user.StaffType, &user.Phone, &user.Aadhar, &user.Address, &user.ImageURL, &user.Branch, &user.ParentID, &user.Active, &user.CreatedAt)
	return user, err
}
func (r *Restaurant) Users(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.Query(ctx, `SELECT id,username,name,password_hash,role,staff_type,phone,aadhar,address,image_url,branch,COALESCE(parent_id,0),active,created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.User{}
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.PasswordHash, &user.Role, &user.StaffType, &user.Phone, &user.Aadhar, &user.Address, &user.ImageURL, &user.Branch, &user.ParentID, &user.Active, &user.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	return result, rows.Err()
}
func (r *Restaurant) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO users(username,name,password_hash,role,staff_type,phone,aadhar,address,image_url,branch,parent_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id,active,created_at`, user.Username, user.Name, user.PasswordHash, user.Role, user.StaffType, user.Phone, user.Aadhar, user.Address, user.ImageURL, user.Branch, nullableID(user.ParentID)).Scan(&user.ID, &user.Active, &user.CreatedAt)
	return user, err
}
func (r *Restaurant) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password_hash=$2 WHERE id=$1`, id, passwordHash)
	return err
}
func (r *Restaurant) SetUserActive(ctx context.Context, id int64, active bool) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, `UPDATE users SET active=$2 WHERE id=$1 RETURNING id,username,name,password_hash,role,staff_type,phone,aadhar,address,image_url,branch,COALESCE(parent_id,0),active,created_at`, id, active).Scan(&user.ID, &user.Username, &user.Name, &user.PasswordHash, &user.Role, &user.StaffType, &user.Phone, &user.Aadhar, &user.Address, &user.ImageURL, &user.Branch, &user.ParentID, &user.Active, &user.CreatedAt)
	return user, err
}
func (r *Restaurant) DeleteUser(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}
func (r *Restaurant) UpdateUser(ctx context.Context, id int64, name, branch string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, `UPDATE users SET name=$2,branch=$3 WHERE id=$1 RETURNING id,username,name,password_hash,role,staff_type,phone,aadhar,address,image_url,branch,COALESCE(parent_id,0),active,created_at`, id, name, branch).Scan(&user.ID, &user.Username, &user.Name, &user.PasswordHash, &user.Role, &user.StaffType, &user.Phone, &user.Aadhar, &user.Address, &user.ImageURL, &user.Branch, &user.ParentID, &user.Active, &user.CreatedAt)
	return user, err
}
func (r *Restaurant) UpdateUserAssignment(ctx context.Context, id int64, name, branch, staffType, phone, aadhar, address, imageURL string, parentID int64) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, `UPDATE users SET name=$2,branch=$3,staff_type=$4,phone=$5,aadhar=$6,address=$7,image_url=$8,parent_id=$9 WHERE id=$1 RETURNING id,username,name,password_hash,role,staff_type,phone,aadhar,address,image_url,branch,COALESCE(parent_id,0),active,created_at`, id, name, branch, staffType, phone, aadhar, address, imageURL, nullableID(parentID)).Scan(&user.ID, &user.Username, &user.Name, &user.PasswordHash, &user.Role, &user.StaffType, &user.Phone, &user.Aadhar, &user.Address, &user.ImageURL, &user.Branch, &user.ParentID, &user.Active, &user.CreatedAt)
	return user, err
}
func (r *Restaurant) PaySalary(ctx context.Context, payment model.SalaryPayment) (model.SalaryPayment, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO salary_payments(user_id,month,amount) VALUES($1,$2,$3) ON CONFLICT(user_id,month) DO UPDATE SET amount=EXCLUDED.amount,paid_at=NOW() RETURNING id,user_id,month,amount,paid_at`, payment.UserID, payment.Month, payment.Amount).Scan(&payment.ID, &payment.UserID, &payment.Month, &payment.Amount, &payment.PaidAt)
	return payment, err
}
func (r *Restaurant) SalaryPayments(ctx context.Context, userID int64) ([]model.SalaryPayment, error) {
	rows, err := r.db.Query(ctx, `SELECT id,user_id,month,amount,paid_at FROM salary_payments WHERE user_id=$1 ORDER BY month DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.SalaryPayment{}
	for rows.Next() {
		var p model.SalaryPayment
		if err := rows.Scan(&p.ID, &p.UserID, &p.Month, &p.Amount, &p.PaidAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *Restaurant) Menu(ctx context.Context) ([]model.MenuItem, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,description,price,emoji FROM menu_items ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.MenuItem{}
	for rows.Next() {
		var item model.MenuItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Emoji); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Restaurant) Specials(ctx context.Context) ([]model.Special, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,description,price,emoji FROM specials WHERE active=true ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Special{}
	for rows.Next() {
		var item model.Special
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Emoji); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Restaurant) Stock(ctx context.Context) ([]model.StockItem, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,quantity,unit,icon FROM stock_items ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.StockItem{}
	for rows.Next() {
		var item model.StockItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &item.Unit, &item.Icon); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Restaurant) Orders(ctx context.Context) ([]model.Order, error) {
	rows, err := r.db.Query(ctx, `SELECT id,customer,date,created_at,items,total,status FROM orders ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Order{}
	for rows.Next() {
		var order model.Order
		var raw []byte
		if err := rows.Scan(&order.ID, &order.Customer, &order.Date, &order.CreatedAt, &raw, &order.Total, &order.Status); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &order.Items); err != nil {
			return nil, err
		}
		result = append(result, order)
	}
	return result, rows.Err()
}
func (r *Restaurant) OrdersByCustomer(ctx context.Context, customer string) ([]model.Order, error) {
	rows, err := r.db.Query(ctx, `SELECT id,customer,date,created_at,items,total,status FROM orders WHERE customer=$1 ORDER BY created_at DESC`, customer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Order{}
	for rows.Next() {
		var order model.Order
		var raw []byte
		if err := rows.Scan(&order.ID, &order.Customer, &order.Date, &order.CreatedAt, &raw, &order.Total, &order.Status); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &order.Items); err != nil {
			return nil, err
		}
		result = append(result, order)
	}
	return result, rows.Err()
}
func (r *Restaurant) CreateOrder(ctx context.Context, order model.Order) (model.Order, error) {
	raw, err := json.Marshal(order.Items)
	if err != nil {
		return order, err
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now().UTC()
	}
	_, err = r.db.Exec(ctx, `INSERT INTO orders(id,customer,date,created_at,items,total,status) VALUES($1,$2,$3,$4,$5,$6,$7)`, order.ID, order.Customer, order.Date, order.CreatedAt, raw, order.Total, order.Status)
	return order, err
}
func (r *Restaurant) UpdateOrder(ctx context.Context, id, status string) (model.Order, error) {
	var order model.Order
	var raw []byte
	err := r.db.QueryRow(ctx, `UPDATE orders SET status=$2 WHERE id=$1 RETURNING id,customer,date,created_at,items,total,status`, id, status).Scan(&order.ID, &order.Customer, &order.Date, &order.CreatedAt, &raw, &order.Total, &order.Status)
	if err != nil {
		return order, err
	}
	err = json.Unmarshal(raw, &order.Items)
	return order, err
}
func (r *Restaurant) DeleteOrder(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM orders WHERE id=$1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
func (r *Restaurant) CreateMenu(ctx context.Context, item model.MenuItem) (model.MenuItem, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO menu_items(name,description,price,emoji) VALUES($1,$2,$3,$4) RETURNING id`, item.Name, item.Description, item.Price, item.Emoji).Scan(&item.ID)
	return item, err
}
func (r *Restaurant) DeleteMenu(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM menu_items WHERE id=$1`, id)
	return err
}
func (r *Restaurant) CreateSpecial(ctx context.Context, item model.Special) (model.Special, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO specials(name,description,price,emoji) VALUES($1,$2,$3,$4) RETURNING id`, item.Name, item.Description, item.Price, item.Emoji).Scan(&item.ID)
	return item, err
}
func (r *Restaurant) DeleteSpecial(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM specials WHERE id=$1`, id)
	return err
}
func (r *Restaurant) CreateStock(ctx context.Context, item model.StockItem) (model.StockItem, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO stock_items(name,quantity,unit,icon) VALUES($1,$2,$3,$4) RETURNING id`, item.Name, item.Quantity, item.Unit, item.Icon).Scan(&item.ID)
	return item, err
}
func (r *Restaurant) UpdateStock(ctx context.Context, id int64, item model.StockItem) (model.StockItem, error) {
	err := r.db.QueryRow(ctx, `UPDATE stock_items SET name=$2,quantity=$3,unit=$4,icon=$5 WHERE id=$1 RETURNING id,name,quantity,unit,icon`, id, item.Name, item.Quantity, item.Unit, item.Icon).Scan(&item.ID, &item.Name, &item.Quantity, &item.Unit, &item.Icon)
	return item, err
}
func (r *Restaurant) DeleteStock(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM stock_items WHERE id=$1`, id)
	return err
}
