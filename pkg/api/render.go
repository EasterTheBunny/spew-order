package api

import "net/http"

// Render implements the render.Renderer interface for use with chi-router
func (ar *BookOrder) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
