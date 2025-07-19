package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"stepikGoWebServices/handlers/utils"
	"stepikGoWebServices/repositories"
)

type RealWorldApi struct {
	http.Handler

	sessionManager *repositories.SessionManager
	userDb         *repositories.UserDbRepository
	articlesDb     *repositories.ArticlesDbRepository
}

func NewRealWorldApi() *RealWorldApi {
	return &RealWorldApi{
		sessionManager: repositories.NewSessionManager(),
		userDb:         repositories.NewUserDbRepository(),
		articlesDb:     repositories.NewArticlesDbRepository(),
	}
}

func (a *RealWorldApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		output *utils.ApiResponse
	)

	switch r.URL.Path {
	case "/api/users":
		output, err = a.handleUsersRequest(r)
	case "/api/users/login":
		output, err = a.handleLoginRequest(r)
	case "/api/user":
		output, err = a.handleUserRequest(r)
	case "/api/user/logout":
		output, err = a.handleLogoutRequest(r)
	case "/api/articles":
		output, err = a.handleArticlesRequest(r)
	default:
		err = utils.ApiError{
			ResponseCode: http.StatusNotFound,
			Err:          "unknown endpoint",
		}
	}

	if err != nil {
		handleApiError(w, err)
	} else if output != nil {
		marshaledResponse, err := json.Marshal(output.Response)
		if err != nil {
			http.Error(w, "failed to marshal output", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(output.Code)
		w.Write(marshaledResponse)
	} else {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
}

func handleApiError(w http.ResponseWriter, err error) {
	var apiErr = utils.ApiError{}
	var ok = errors.As(err, &apiErr)

	var jsonError = map[string]string{
		"error": apiErr.Error(),
	}

	marshalledJson, err := json.Marshal(jsonError)
	if err != nil {
		http.Error(w, "failed to marshal output", http.StatusInternalServerError)
		return
	}

	if ok {
		w.WriteHeader(apiErr.ResponseCode)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(marshalledJson)
}
