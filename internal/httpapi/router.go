package httpapi

import (
	"log/slog"
	"net/http"
	"time"
	"github.com/jhatkaz/restaurant-console/internal/service"
)

func NewRouter(app *service.Restaurant, cors string, logger *slog.Logger) http.Handler {
	h := NewHandler(app)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health); mux.HandleFunc("GET /restaurant/bootstrap", h.bootstrap); mux.HandleFunc("GET /menu", h.menu); mux.HandleFunc("POST /menu", h.createMenu); mux.HandleFunc("DELETE /menu/{id}", h.deleteMenu); mux.HandleFunc("GET /specials/today", h.specials); mux.HandleFunc("POST /specials", h.createSpecial); mux.HandleFunc("DELETE /specials/{id}", h.deleteSpecial); mux.HandleFunc("GET /stock", h.stock); mux.HandleFunc("POST /stock", h.createStock); mux.HandleFunc("PATCH /stock/{id}", h.updateStock); mux.HandleFunc("DELETE /stock/{id}", h.deleteStock); mux.HandleFunc("GET /orders", h.orders); mux.HandleFunc("POST /orders", h.createOrder); mux.HandleFunc("PATCH /orders/{id}", h.updateOrder); mux.HandleFunc("DELETE /orders/{id}", h.deleteOrder)
	return accessLog(cors, logger, mux)
}

func accessLog(cors string, logger *slog.Logger, next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ w.Header().Set("Access-Control-Allow-Origin",cors);w.Header().Set("Access-Control-Allow-Headers","Content-Type, Authorization");w.Header().Set("Access-Control-Allow-Methods","GET, POST, PATCH, DELETE, OPTIONS");if r.Method==http.MethodOptions{w.WriteHeader(http.StatusNoContent);return};started:=time.Now();next.ServeHTTP(w,r);logger.Info("http request","method",r.Method,"path",r.URL.Path,"duration",time.Since(started).String()) }) }
