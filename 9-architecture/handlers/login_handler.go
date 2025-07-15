package handlers

import (
	"encoding/json"
	"net/http"
	"stepikGoWebServices/handlers/utils"
)

func (a *RealWorldApi) handleLoginRequest(r *http.Request) (*utils.ApiResponse, error) {
	switch r.Method {
	case http.MethodPost:
		return a.loginUser(r)
	default:
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "bad method",
		}
	}
}

func (a *RealWorldApi) loginUser(r *http.Request) (*utils.ApiResponse, error) {
	var req struct {
		User struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "failed to decode body",
		}
	}

	var user = a.userDb.GetUserByEmail(req.User.Email)
	if user == nil || user.Password != req.User.Password {
		return nil, utils.ApiError{
			ResponseCode: http.StatusUnauthorized,
			Err:          "forbidden",
		}
	}

	user.Token = a.sessionManager.AddSession(user.Id)

	var response = map[string]interface{}{
		"user": user,
	}

	return &utils.ApiResponse{
		Code:     http.StatusOK,
		Response: response,
	}, nil
}
