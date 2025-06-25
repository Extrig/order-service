package handlers

import (
	"encoding/json"
	"github.com/Extrig/order-service/internal/cache"
	"github.com/Extrig/order-service/internal/db"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

func GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["uid"]
	log.Info().Msgf("Get order with id: %s", orderID)
	// Сначала ищем в кеше
	order, ok := cache.Get(orderID)
	if !ok {
		// Если нет — пробуем из БД
		var err error
		order, err = db.GetOrderById(orderID)
		if err != nil {
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}
		// Кладем в кеш
		cache.Set(order)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
