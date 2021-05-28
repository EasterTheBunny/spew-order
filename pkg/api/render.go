package api

import "net/http"

// Render implements the render.Renderer interface for use with chi-router
func (ar *BookOrder) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render implements the render.Renderer interface for use with chi-router
func (a *Account) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render implements the render.Renderer interface for use with chi-router
func (b *BalanceItem) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render implements the render.Renderer interface for use with chi-router
func (b *BalanceList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
