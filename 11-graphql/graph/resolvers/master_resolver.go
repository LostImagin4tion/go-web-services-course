package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"stepikGoWebServices/graph/model"
	"strings"
)

type TestData struct {
	Catalogs []CatalogData `json:"catalogs"`
	Items    []ItemData    `json:"items"`
	Sellers  []SellerData  `json:"sellers"`
}

type CatalogData struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id"`
}

type ItemData struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CatalogID int    `json:"catalog_id"`
	SellerID  int    `json:"seller_id"`
	InStock   int    `json:"in_stock"`
}

type SellerData struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Deals int    `json:"deals"`
}

type RegistrationRequest struct {
	User struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	} `json:"user"`
}

type RegistrationResponse struct {
	Token string `json:"token"`
}

type Resolver struct {
	Data       TestData
	UserCarts  map[string][]model.CartItem // userID -> cart items
	UserTokens map[string]string           // token -> userID
}

func (res *Resolver) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := fmt.Sprintf("token_%s_%d", req.User.Username, len(res.UserTokens))
	res.UserTokens[token] = req.User.Username

	response := map[string]interface{}{
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (res *Resolver) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Token ") {
			token := strings.TrimPrefix(authHeader, "Token ")
			if userID, exists := res.UserTokens[token]; exists {
				ctx := context.WithValue(r.Context(), "user", userID)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
