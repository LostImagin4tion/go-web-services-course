package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	errTest    = errors.New("testing")
	httpClient = &http.Client{
		Timeout: time.Second,
	}
)

type User struct {
	fmt.Stringer

	Id     int
	Name   string
	Age    int
	About  string
	Gender string
}

func (u User) String() string {
	return fmt.Sprintf(
		"User(id=%v, name=%v, age=%v gender=%v),",
		u.Id,
		u.Name,
		u.Age,
		u.Gender,
	)
}

type SearchResponse struct {
	Users    []User
	NextPage bool
}

type SearchErrorResponse struct {
	Error string
}

const (
	OrderByDesc = -1
	OrderByAsIs = 0
	OrderByAsc  = 1

	OrderFieldId   = "Id"
	OrderFieldAge  = "Age"
	OrderFieldName = "Name"

	ErrorBadOrderField = "OrderField invalid"
	ErrorBadOrderBy    = "OrderBy invalid"
)

var (
	validOrderFieldValues = []string{
		OrderFieldId,
		OrderFieldAge,
		OrderFieldName,
	}
)

type SearchRequest struct {
	Limit      int
	Offset     int    // Можно учесть после сортировки
	Query      string // подстрока в 1 из полей
	OrderField string
	OrderBy    int
}

type SearchClient struct {
	AccessToken string
	URL         string
}

// FindUsers отправляет запрос во внешнюю систему, которая непосредственно ищет пользователей
func (search *SearchClient) FindUsers(req SearchRequest) (*SearchResponse, error) {

	searchParams := url.Values{}

	if req.Limit < 0 {
		return nil, fmt.Errorf("limit must be > 0")
	}
	if req.Limit > 25 {
		req.Limit = 25
	}
	if req.Offset < 0 {
		return nil, fmt.Errorf("offset must be > 0")
	}

	//нужно для получения следующей записи, на основе которой мы скажем - можно показать переключатель следующей страницы или нет
	req.Limit++

	searchParams.Add("limit", strconv.Itoa(req.Limit))
	searchParams.Add("offset", strconv.Itoa(req.Offset))
	searchParams.Add("query", req.Query)
	searchParams.Add("order_field", req.OrderField)
	searchParams.Add("order_by", strconv.Itoa(req.OrderBy))

	searchRequest, err := http.NewRequest(
		"GET",
		search.URL+"?"+searchParams.Encode(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	searchRequest.Header.Add("AccessToken", search.AccessToken)

	resp, err := httpClient.Do(searchRequest)
	if err != nil {
		var netError net.Error
		if errors.As(err, &netError) && netError.Timeout() {
			return nil, fmt.Errorf("timeout for %s", searchParams.Encode())
		}
		return nil, fmt.Errorf("unknown error %s", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("bad AccessToken")

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("SearchServer fatal error")

	case http.StatusBadRequest:
		errResp := SearchErrorResponse{}
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return nil, fmt.Errorf("cant unpack error json: %s", err)
		}
		if errResp.Error == ErrorBadOrderField {
			return nil, fmt.Errorf("OrderField %s invalid", req.OrderField)
		}
		return nil, fmt.Errorf("unknown bad request error: %s", errResp.Error)
	}

	data := make([]User, 0)
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("cant unpack result json: %s", err)
	}

	result := SearchResponse{}
	if len(data) == req.Limit {
		result.NextPage = true
		result.Users = data[0 : len(data)-1]
	} else {
		result.Users = data[0:]
	}

	return &result, err
}
