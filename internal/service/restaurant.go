package service

import (
	"context"
	"github.com/jhatkaz/restaurant-console/internal/model"
	"github.com/jhatkaz/restaurant-console/internal/repository"
)

type Restaurant struct { repo *repository.Restaurant }
func NewRestaurant(repo *repository.Restaurant) *Restaurant { return &Restaurant{repo: repo} }
func (s *Restaurant) Bootstrap(ctx context.Context) (model.Bootstrap,error) { menu,e:=s.repo.Menu(ctx);if e!=nil{return model.Bootstrap{},e}; specials,e:=s.repo.Specials(ctx);if e!=nil{return model.Bootstrap{},e};orders,e:=s.repo.Orders(ctx);if e!=nil{return model.Bootstrap{},e};stock,e:=s.repo.Stock(ctx);if e!=nil{return model.Bootstrap{},e};return model.Bootstrap{Menu:menu,Specials:specials,Orders:orders,Stock:stock},nil }
func (s *Restaurant) Menu(ctx context.Context)([]model.MenuItem,error){return s.repo.Menu(ctx)}
func (s *Restaurant) Specials(ctx context.Context)([]model.Special,error){return s.repo.Specials(ctx)}
func (s *Restaurant) Stock(ctx context.Context)([]model.StockItem,error){return s.repo.Stock(ctx)}
func (s *Restaurant) Orders(ctx context.Context)([]model.Order,error){return s.repo.Orders(ctx)}
func (s *Restaurant) CreateOrder(ctx context.Context,o model.Order)(model.Order,error){return s.repo.CreateOrder(ctx,o)}
func (s *Restaurant) UpdateOrder(ctx context.Context,id,status string)(model.Order,error){return s.repo.UpdateOrder(ctx,id,status)}
func (s *Restaurant) DeleteOrder(ctx context.Context,id string)error{return s.repo.DeleteOrder(ctx,id)}
func (s *Restaurant) CreateMenu(ctx context.Context,i model.MenuItem)(model.MenuItem,error){return s.repo.CreateMenu(ctx,i)}
func (s *Restaurant) DeleteMenu(ctx context.Context,id int64)error{return s.repo.DeleteMenu(ctx,id)}
func (s *Restaurant) CreateSpecial(ctx context.Context,i model.Special)(model.Special,error){return s.repo.CreateSpecial(ctx,i)}
func (s *Restaurant) DeleteSpecial(ctx context.Context,id int64)error{return s.repo.DeleteSpecial(ctx,id)}
func (s *Restaurant) CreateStock(ctx context.Context,i model.StockItem)(model.StockItem,error){return s.repo.CreateStock(ctx,i)}
func (s *Restaurant) UpdateStock(ctx context.Context,id int64,i model.StockItem)(model.StockItem,error){return s.repo.UpdateStock(ctx,id,i)}
func (s *Restaurant) DeleteStock(ctx context.Context,id int64)error{return s.repo.DeleteStock(ctx,id)}
