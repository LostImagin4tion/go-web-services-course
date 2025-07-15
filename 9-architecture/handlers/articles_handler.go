package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"stepikGoWebServices/handlers/entities"
	"stepikGoWebServices/handlers/utils"
)

func (a *RealWorldApi) handleArticlesRequest(r *http.Request) (*utils.ApiResponse, error) {
	switch r.Method {
	case http.MethodGet:
		return a.getArticles(r)
	case http.MethodPost:
		return a.createArticle(r)
	default:
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "bad method",
		}
	}
}

func (a *RealWorldApi) createArticle(r *http.Request) (*utils.ApiResponse, error) {
	user := a.getAuthenticatedUser(r)
	if user == nil {
		return nil, utils.ApiError{
			ResponseCode: http.StatusUnauthorized,
			Err:          "forbidden",
		}
	}

	var req struct {
		Article struct {
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Body        string   `json:"body"`
			TagList     []string `json:"tagList"`
		} `json:"article"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, utils.ApiError{
			ResponseCode: http.StatusBadRequest,
			Err:          "failed to parse body",
		}
	}

	var article = a.articlesDb.AddArticle(
		req.Article.Title,
		req.Article.Description,
		req.Article.Body,
		req.Article.TagList,
		*user,
	)

	var response = map[string]interface{}{
		"article": article,
	}

	return &utils.ApiResponse{
		Code:     http.StatusCreated,
		Response: response,
	}, nil
}

func (a *RealWorldApi) getArticles(r *http.Request) (*utils.ApiResponse, error) {
	query := r.URL.Query()
	author := query.Get("author")
	tag := query.Get("tag")

	var articles = a.articlesDb.GetArticles(author, tag)

	slices.SortFunc(
		articles,
		func(a, b entities.Article) int {
			switch {
			case a.CreatedAt.Before(b.CreatedAt):
				return -1
			case a.CreatedAt.After(b.CreatedAt):
				return 1
			default:
				return 0
			}
		},
	)

	var response = map[string]interface{}{
		"articles":      articles,
		"articlesCount": len(articles),
	}

	return &utils.ApiResponse{
		Code:     http.StatusOK,
		Response: response,
	}, nil
}
