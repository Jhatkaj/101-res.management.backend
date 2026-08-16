package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jhatkaz/restaurant-console/internal/model"
)

type Restaurant struct { db *pgxpool.Pool }
func New(db *pgxpool.Pool) *Restaurant { return &Restaurant{db: db} }

func (r *Restaurant) Menu(ctx context.Context) ([]model.MenuItem, error) { rows, err := r.db.Query(ctx, `SELECT id,name,description,price,emoji FROM menu_items ORDER BY id`); if err != nil { return nil, err }; defer rows.Close(); result := []model.MenuItem{}; for rows.Next() { var item model.MenuItem; if err := rows.Scan(&item.ID,&item.Name,&item.Description,&item.Price,&item.Emoji); err != nil { return nil, err }; result = append(result,item) }; return result, rows.Err() }
func (r *Restaurant) Specials(ctx context.Context) ([]model.Special, error) { rows, err := r.db.Query(ctx, `SELECT id,name,description,price,emoji FROM specials WHERE active=true ORDER BY created_at DESC`); if err != nil { return nil, err }; defer rows.Close(); result := []model.Special{}; for rows.Next() { var item model.Special; if err := rows.Scan(&item.ID,&item.Name,&item.Description,&item.Price,&item.Emoji); err != nil { return nil, err }; result = append(result,item) }; return result, rows.Err() }
func (r *Restaurant) Stock(ctx context.Context) ([]model.StockItem, error) { rows, err := r.db.Query(ctx, `SELECT id,name,quantity,unit,icon FROM stock_items ORDER BY id`); if err != nil { return nil, err }; defer rows.Close(); result := []model.StockItem{}; for rows.Next() { var item model.StockItem; if err := rows.Scan(&item.ID,&item.Name,&item.Quantity,&item.Unit,&item.Icon); err != nil { return nil, err }; result = append(result,item) }; return result, rows.Err() }
func (r *Restaurant) Orders(ctx context.Context) ([]model.Order, error) { rows, err := r.db.Query(ctx, `SELECT id,customer,date,created_at,items,total,status FROM orders ORDER BY created_at DESC`); if err != nil { return nil, err }; defer rows.Close(); result := []model.Order{}; for rows.Next() { var order model.Order; var raw []byte; if err := rows.Scan(&order.ID,&order.Customer,&order.Date,&order.CreatedAt,&raw,&order.Total,&order.Status); err != nil { return nil, err }; if err := json.Unmarshal(raw,&order.Items); err != nil { return nil, err }; result = append(result,order) }; return result, rows.Err() }
func (r *Restaurant) CreateOrder(ctx context.Context, order model.Order) (model.Order, error) { raw, err := json.Marshal(order.Items); if err != nil { return order, err }; if order.CreatedAt.IsZero() { order.CreatedAt = time.Now().UTC() }; _, err = r.db.Exec(ctx, `INSERT INTO orders(id,customer,date,created_at,items,total,status) VALUES($1,$2,$3,$4,$5,$6,$7)`, order.ID,order.Customer,order.Date,order.CreatedAt,raw,order.Total,order.Status); return order, err }
func (r *Restaurant) UpdateOrder(ctx context.Context, id, status string) (model.Order, error) { var order model.Order; var raw []byte; err := r.db.QueryRow(ctx, `UPDATE orders SET status=$2 WHERE id=$1 RETURNING id,customer,date,created_at,items,total,status`, id,status).Scan(&order.ID,&order.Customer,&order.Date,&order.CreatedAt,&raw,&order.Total,&order.Status); if err != nil { return order, err }; err = json.Unmarshal(raw,&order.Items); return order, err }
func (r *Restaurant) DeleteOrder(ctx context.Context, id string) error { tag, err := r.db.Exec(ctx, `DELETE FROM orders WHERE id=$1`, id); if err == nil && tag.RowsAffected() == 0 { return pgx.ErrNoRows }; return err }
func (r *Restaurant) CreateMenu(ctx context.Context, item model.MenuItem) (model.MenuItem,error) { err := r.db.QueryRow(ctx, `INSERT INTO menu_items(name,description,price,emoji) VALUES($1,$2,$3,$4) RETURNING id`, item.Name,item.Description,item.Price,item.Emoji).Scan(&item.ID); return item,err }
func (r *Restaurant) DeleteMenu(ctx context.Context,id int64) error { _,err:=r.db.Exec(ctx,`DELETE FROM menu_items WHERE id=$1`,id);return err }
func (r *Restaurant) CreateSpecial(ctx context.Context,item model.Special)(model.Special,error){err:=r.db.QueryRow(ctx,`INSERT INTO specials(name,description,price,emoji) VALUES($1,$2,$3,$4) RETURNING id`,item.Name,item.Description,item.Price,item.Emoji).Scan(&item.ID);return item,err}
func (r *Restaurant) DeleteSpecial(ctx context.Context,id int64)error{_,err:=r.db.Exec(ctx,`DELETE FROM specials WHERE id=$1`,id);return err}
func (r *Restaurant) CreateStock(ctx context.Context,item model.StockItem)(model.StockItem,error){err:=r.db.QueryRow(ctx,`INSERT INTO stock_items(name,quantity,unit,icon) VALUES($1,$2,$3,$4) RETURNING id`,item.Name,item.Quantity,item.Unit,item.Icon).Scan(&item.ID);return item,err}
func (r *Restaurant) UpdateStock(ctx context.Context,id int64,item model.StockItem)(model.StockItem,error){err:=r.db.QueryRow(ctx,`UPDATE stock_items SET name=$2,quantity=$3,unit=$4,icon=$5 WHERE id=$1 RETURNING id,name,quantity,unit,icon`,id,item.Name,item.Quantity,item.Unit,item.Icon).Scan(&item.ID,&item.Name,&item.Quantity,&item.Unit,&item.Icon);return item,err}
func (r *Restaurant) DeleteStock(ctx context.Context,id int64)error{_,err:=r.db.Exec(ctx,`DELETE FROM stock_items WHERE id=$1`,id);return err}
