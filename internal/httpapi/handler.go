package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jhatkaz/restaurant-console/internal/model"
	"github.com/jhatkaz/restaurant-console/internal/service"
)

type Handler struct { app *service.Restaurant }
func NewHandler(app *service.Restaurant) *Handler { return &Handler{app:app} }
func (h *Handler) health(w http.ResponseWriter,_ *http.Request){writeJSON(w,http.StatusOK,map[string]string{"status":"ok"})}
func (h *Handler) bootstrap(w http.ResponseWriter,r *http.Request){v,e:=h.app.Bootstrap(r.Context());respond(w,v,e)}
func (h *Handler) menu(w http.ResponseWriter,r *http.Request){v,e:=h.app.Menu(r.Context());respond(w,v,e)}
func (h *Handler) specials(w http.ResponseWriter,r *http.Request){v,e:=h.app.Specials(r.Context());respond(w,v,e)}
func (h *Handler) stock(w http.ResponseWriter,r *http.Request){v,e:=h.app.Stock(r.Context());respond(w,v,e)}
func (h *Handler) orders(w http.ResponseWriter,r *http.Request){v,e:=h.app.Orders(r.Context());respond(w,v,e)}
func (h *Handler) createOrder(w http.ResponseWriter,r *http.Request){var v model.Order;if !decode(w,r,&v){return};if v.ID==""||v.Customer==""||len(v.Items)==0{errorJSON(w,http.StatusBadRequest,"id, customer, and items are required");return};saved,e:=h.app.CreateOrder(r.Context(),v);respondStatus(w,saved,e,http.StatusCreated)}
func (h *Handler) updateOrder(w http.ResponseWriter,r *http.Request){var body struct{Status string `json:"status"`};if !decode(w,r,&body){return};if body.Status==""{errorJSON(w,http.StatusBadRequest,"status is required");return};v,e:=h.app.UpdateOrder(r.Context(),r.PathValue("id"),body.Status);respond(w,v,e)}
func (h *Handler) deleteOrder(w http.ResponseWriter,r *http.Request){e:=h.app.DeleteOrder(r.Context(),r.PathValue("id"));if e!=nil{errorJSON(w,http.StatusNotFound,"order not found");return};w.WriteHeader(http.StatusNoContent)}
func (h *Handler) createMenu(w http.ResponseWriter,r *http.Request){var v model.MenuItem;if !decode(w,r,&v){return};saved,e:=h.app.CreateMenu(r.Context(),v);respondStatus(w,saved,e,http.StatusCreated)}
func (h *Handler) deleteMenu(w http.ResponseWriter,r *http.Request){id,e:=idParam(r);if e!=nil{errorJSON(w,400,"invalid id");return};if e=h.app.DeleteMenu(r.Context(),id);e!=nil{errorJSON(w,404,"menu item not found");return};w.WriteHeader(204)}
func (h *Handler) createSpecial(w http.ResponseWriter,r *http.Request){var v model.Special;if !decode(w,r,&v){return};saved,e:=h.app.CreateSpecial(r.Context(),v);respondStatus(w,saved,e,201)}
func (h *Handler) deleteSpecial(w http.ResponseWriter,r *http.Request){id,e:=idParam(r);if e!=nil{errorJSON(w,400,"invalid id");return};if e=h.app.DeleteSpecial(r.Context(),id);e!=nil{errorJSON(w,404,"special not found");return};w.WriteHeader(204)}
func (h *Handler) createStock(w http.ResponseWriter,r *http.Request){var v model.StockItem;if !decode(w,r,&v){return};saved,e:=h.app.CreateStock(r.Context(),v);respondStatus(w,saved,e,201)}
func (h *Handler) updateStock(w http.ResponseWriter,r *http.Request){id,e:=idParam(r);if e!=nil{errorJSON(w,400,"invalid id");return};var v model.StockItem;if !decode(w,r,&v){return};saved,e:=h.app.UpdateStock(r.Context(),id,v);respond(w,saved,e)}
func (h *Handler) deleteStock(w http.ResponseWriter,r *http.Request){id,e:=idParam(r);if e!=nil{errorJSON(w,400,"invalid id");return};if e=h.app.DeleteStock(r.Context(),id);e!=nil{errorJSON(w,404,"stock item not found");return};w.WriteHeader(204)}
func idParam(r *http.Request)(int64,error){return strconv.ParseInt(strings.TrimSpace(r.PathValue("id")),10,64)}
func decode(w http.ResponseWriter,r *http.Request,v any)bool{defer r.Body.Close();if e:=json.NewDecoder(r.Body).Decode(v);e!=nil{errorJSON(w,400,"invalid JSON body");return false};return true}
func respond(w http.ResponseWriter,v any,e error){if e!=nil{errorJSON(w,500,"internal server error");return};writeJSON(w,200,v)}
func respondStatus(w http.ResponseWriter,v any,e error,status int){if e!=nil{errorJSON(w,500,"internal server error");return};writeJSON(w,status,v)}
func writeJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(v)}
func errorJSON(w http.ResponseWriter,status int,message string){writeJSON(w,status,map[string]string{"error":message})}
