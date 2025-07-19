package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"stepikGoWebServices/handlers/utils"
)

func (a *RealWorldApi) handleUsersRequest(r *http.Request) (*utils.ApiResponse, error) {
	switch r.Method {
	case http.MethodPost:
		return a.registerUser(r)
	default:
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "bad method",
		}
	}
}

func (a *RealWorldApi) registerUser(r *http.Request) (*utils.ApiResponse, error) {
	var req struct {
		User struct {
			Email    string `json:"email"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"user"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "failed to decode body json",
		}
	}

	if a.userDb.IsUserExists(req.User.Email) {
		return nil, utils.ApiError{
			ResponseCode: http.StatusConflict,
			Err:          "user already exists",
		}
	}

	var user = a.userDb.AddUser(
		req.User.Email,
		req.User.Username,
		req.User.Password,
	)
	user.Token = a.sessionManager.AddSession(user.Id)

	var response = map[string]any{
		"user": user,
	}

	log.Printf("HELLO user %v\n", user)

	return &utils.ApiResponse{
		Code:     http.StatusCreated,
		Response: response,
	}, nil
}
