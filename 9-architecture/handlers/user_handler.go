package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stepikGoWebServices/handlers/entities"
	"stepikGoWebServices/handlers/utils"
	"strings"
	"time"
)

func (a *RealWorldApi) handleUserRequest(r *http.Request) (*utils.ApiResponse, error) {
	switch r.Method {
	case http.MethodGet:
		return a.getCurrentUser(r)
	case http.MethodPut:
		return a.updateUser(r)
	default:
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "bad method",
		}
	}
}

func (a *RealWorldApi) getCurrentUser(r *http.Request) (*utils.ApiResponse, error) {
	user := a.getAuthenticatedUser(r)
	fmt.Printf("HELLO current user %v\n", user)
	if user == nil {
		return nil, utils.ApiError{
			ResponseCode: http.StatusUnauthorized,
			Err:          "forbidden",
		}
	}

	return &utils.ApiResponse{
		Code: http.StatusOK,
		Response: map[string]any{
			"user": user,
		},
	}, nil
}

func (a *RealWorldApi) updateUser(r *http.Request) (*utils.ApiResponse, error) {
	currUser := a.getAuthenticatedUser(r)
	if currUser == nil {
		return nil, utils.ApiError{
			ResponseCode: http.StatusUnauthorized,
			Err:          "forbidden",
		}
	}

	var req struct {
		User struct {
			Email string `json:"email"`
			Bio   string `json:"bio"`
		} `json:"user"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "failed to parse body",
		}
	}

	var newUser = &entities.User{
		Id:        currUser.Id,
		Email:     req.User.Email,
		Username:  currUser.Username,
		Password:  currUser.Password,
		Bio:       req.User.Bio,
		Image:     currUser.Image,
		CreatedAt: currUser.CreatedAt,
		UpdatedAt: time.Now(),
		Token:     currUser.Token,
		Following: currUser.Following,
	}

	a.userDb.UpdateUser(currUser, newUser)
	a.sessionManager.UpdateSession(newUser.Token, newUser.Id)

	fmt.Printf("HELLO old user %v\n", currUser)
	fmt.Printf("HELLO new user %v\n", newUser)

	var response = map[string]interface{}{
		"user": newUser,
	}

	return &utils.ApiResponse{
		Code:     http.StatusOK,
		Response: response,
	}, nil
}

func (a *RealWorldApi) getAuthenticatedUser(r *http.Request) *entities.User {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil
	}

	if !strings.HasPrefix(auth, "Token ") {
		return nil
	}

	token := strings.TrimPrefix(auth, "Token ")

	var userId = a.sessionManager.GetAuthenticatedUser(token)
	return a.userDb.GetUserById(userId)
}
