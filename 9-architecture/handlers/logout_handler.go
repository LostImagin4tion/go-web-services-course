package handlers

import (
	"net/http"
	"stepikGoWebServices/handlers/utils"
)

func (a *RealWorldApi) handleLogoutRequest(r *http.Request) (*utils.ApiResponse, error) {
	switch r.Method {
	case http.MethodPost:
		return a.logoutUser(r)
	default:
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "bad method",
		}
	}
}

func (a *RealWorldApi) logoutUser(r *http.Request) (*utils.ApiResponse, error) {
	user := a.getAuthenticatedUser(r)
	if user == nil {
		return nil, utils.ApiError{
			ResponseCode: http.StatusUnauthorized,
			Err:          "forbidden",
		}
	}

	a.sessionManager.DeleteSession(user.Token)

	return &utils.ApiResponse{
		Code:     http.StatusOK,
		Response: make(map[string]any),
	}, nil
}
