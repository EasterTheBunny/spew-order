package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/easterthebunny/spew-order/pkg/queue"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type RESTHandler struct {
	queue *queue.OrderQueue
}

func NewRESTHandler(q *queue.OrderQueue) *RESTHandler {
	return &RESTHandler{queue: q}
}

// PostOrder publishes a message to Pub/Sub. PublishMessage only works
// with topics that already exist.
func (h *RESTHandler) PostOrder(w http.ResponseWriter, r *http.Request) {
	var or types.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&or); err != nil {
		log.Printf("json.NewDecoder: %v", err)
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	// TODO: get account and add it to request

	id, err := h.queue.PublishOrderRequest(r.Context(), or)
	if err != nil {
		log.Printf("topic(%s).Publish.Get: %v", queue.OrderTopic, err)
		http.Error(w, "Error publishing message", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Message published: %v", id)
}
