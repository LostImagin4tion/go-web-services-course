package main

import (
	"cmp"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	usersDatabasePath  = "./dataset.xml"
	accessTokenCorrect = "accessToken"
)

var (
	usersDb = make([]User, 0)
)

type databaseData struct {
	XMLName xml.Name      `xml:"root"`
	XMLRows []databaseRow `xml:"row"`
}

type databaseRow struct {
	XMLName   xml.Name `xml:"row"`
	Id        int      `xml:"id"`
	FirstName string   `xml:"first_name"`
	LastName  string   `xml:"last_name"`
	Age       int      `xml:"age"`
	About     string   `xml:"about"`
	Gender    string   `xml:"gender"`
}

func LoadUserDatabase() {
	var file, err = os.Open(usersDatabasePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var dbData databaseData
	var data, _ = io.ReadAll(file)

	err = xml.Unmarshal(data, &dbData)
	if err != nil {
		panic("Failed to unmarshal database")
	}

	for _, row := range dbData.XMLRows {
		usersDb = append(
			usersDb,
			User{
				Id:     row.Id,
				Name:   fmt.Sprintf("%v %v", row.FirstName, row.LastName),
				Age:    row.Age,
				About:  row.About,
				Gender: row.Gender,
			},
		)
	}

	fmt.Printf("Decoded %v users: %v", len(usersDb), usersDb)
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	var accessToken = r.Header.Get("AccessToken")
	if accessToken != accessTokenCorrect {
		sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var searchRequest, err = parseSearchRequest(r.URL)
	if err != nil {
		sendError(
			w,
			fmt.Sprintf("Failed to parse search request %v", err),
			http.StatusBadRequest,
		)
		return
	}

	if len(searchRequest.OrderField) != 0 && slices.Contains(validOrderFieldValues, searchRequest.OrderField) {
		sendError(w, ErrorBadOrderField, http.StatusBadRequest)
		return
	}
	if searchRequest.OrderBy < -1 || searchRequest.OrderBy > 1 {
		sendError(w, ErrorBadOrderBy, http.StatusBadRequest)
		return
	}
	if searchRequest.Limit < 0 {
		sendError(w, "Limit must be >= 0", http.StatusBadRequest)
	}
	if searchRequest.Offset < 0 {
		sendError(w, "Offset must be >= 0", http.StatusBadRequest)
	}

	var query = searchRequest.Query
	var queriedUsers = make([]User, 0)

	if len(query) == 0 {
		queriedUsers = make([]User, len(usersDb))
		copy(queriedUsers, usersDb)
	} else {
		for _, user := range usersDb {
			if strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
				queriedUsers = append(queriedUsers, user)
			}
		}
	}

	if searchRequest.OrderBy != OrderByAsIs {
		performSort(&queriedUsers, searchRequest.OrderField, searchRequest.OrderBy)
	}

	var result []User

	var from = searchRequest.Offset
	if from > len(queriedUsers)-1 {
		result = make([]User, 0)
	} else {
		var to = min(
			searchRequest.Offset+searchRequest.Limit,
			len(queriedUsers),
		)
		result = queriedUsers[from:to]
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		sendError(
			w,
			fmt.Sprintf("Error happened while parsing queried users: %v", err),
			http.StatusInternalServerError,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func sendError(w http.ResponseWriter, str string, status int) {
	j, err := json.Marshal(SearchErrorResponse{str})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(j)
}

func parseSearchRequest(url *url.URL) (*SearchRequest, error) {
	var queryParams = url.Query()

	var err error

	var limitStr = queryParams.Get("limit")
	var limit int
	if len(limitStr) != 0 {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("cant convert limit value to int %v", err)
		}
	}

	var offsetStr = queryParams.Get("offset")
	var offset int
	if len(offsetStr) != 0 {
		offset, err = strconv.Atoi(queryParams.Get("offset"))
		if err != nil {
			return nil, fmt.Errorf("cant convert offset value to int %v", err)
		}
	}

	var orderByStr = queryParams.Get("order_by")
	var orderBy int
	if len(orderByStr) != 0 {
		orderBy, err = strconv.Atoi(orderByStr)
		if err != nil {
			return nil, fmt.Errorf("cant convert order_by value to int %v", err)
		}
	}

	var query = queryParams.Get("query")
	var orderField = queryParams.Get("order_field")

	return &SearchRequest{
		Limit:      limit,
		Offset:     offset,
		Query:      query,
		OrderField: orderField,
		OrderBy:    orderBy,
	}, nil
}

func performSort(users *[]User, orderField string, orderBy int) {
	slices.SortFunc(
		*users,
		func(a, b User) int {
			switch {
			case orderField == OrderFieldName || len(orderField) == 0:
				var compare = cmp.Compare(a.Name, b.Name)
				return orderBy * compare

			case orderField == OrderFieldId:
				var compare = cmp.Compare(a.Id, b.Id)
				return orderBy * compare

			case orderField == OrderFieldAge:
				var compare = cmp.Compare(a.Age, b.Age)
				return orderBy * compare

			default:
				return -1
			}
		},
	)
}
