package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
)

type TestCase struct {
	request            SearchRequest
	expectedResponse   *SearchResponse
	expectedError      error
	shouldCompareUsers bool
}

func (test *TestCase) validate(
	t *testing.T,
	response *SearchResponse,
	err error,
) {
	if err != nil {
		if test.expectedError == nil {
			t.Errorf("Unexpected error occured: %v", err)
		} else {
			var errStr = err.Error()
			var expectedErrStr = test.expectedError.Error()

			if expectedErrStr != errStr && !strings.Contains(errStr, expectedErrStr) {
				t.Errorf("Unexpected error occured: %v,\nbut expected %v", err, test.expectedError)
			}
		}
	} else {
		if len(response.Users) != len(test.expectedResponse.Users) {
			t.Errorf(
				"Invalid number of users. Got %v, expected %v",
				len(response.Users),
				len(test.expectedResponse.Users),
			)
		}
		if test.shouldCompareUsers && !slices.Equal(response.Users, test.expectedResponse.Users) {
			t.Errorf(
				"Got wrong response. Got %v\nExpected %v",
				response.Users,
				test.expectedResponse.Users,
			)
		}
	}
}

type TestEnv struct {
	Server *httptest.Server
	Client SearchClient
}

func init() {
	LoadUserDatabase()
}

func InitTestEnv(accessToken string) *TestEnv {
	var server = httptest.NewServer(http.HandlerFunc(SearchServer))
	var client = SearchClient{
		AccessToken: accessToken,
		URL:         server.URL,
	}
	return &TestEnv{
		Server: server,
		Client: client,
	}
}

func InitTestEnvWithUrl(accessToken string, url string) *TestEnv {
	var server = httptest.NewServer(http.HandlerFunc(SearchServer))
	var client = SearchClient{
		AccessToken: accessToken,
		URL:         url,
	}
	return &TestEnv{
		Server: server,
		Client: client,
	}
}

func InitTestEnvWithHandler(
	accessToken string,
	handler func(http.ResponseWriter, *http.Request),
) *TestEnv {
	var server = httptest.NewServer(http.HandlerFunc(handler))
	var client = SearchClient{
		AccessToken: accessToken,
		URL:         server.URL,
	}
	return &TestEnv{
		Server: server,
		Client: client,
	}
}

func TestFindUser(t *testing.T) {
	var testCases = []TestCase{
		{
			request: SearchRequest{
				Limit: 1,
				Query: "Everett",
			},
			expectedResponse: &SearchResponse{
				Users: []User{
					{
						Id:     3,
						Name:   "Everett Dillard",
						Age:    27,
						About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
		},
		{
			request: SearchRequest{
				Limit:      2,
				Query:      "Dillard",
				OrderField: OrderFieldName,
				OrderBy:    OrderByAsc,
			},
			expectedResponse: &SearchResponse{
				Users: []User{
					{
						Id:     17,
						Name:   "Dillard Mccoy",
						Age:    36,
						About:  "Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.",
						Gender: "male",
					},
					{
						Id:     3,
						Name:   "Everett Dillard",
						Age:    27,
						About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
		},
		{
			request: SearchRequest{
				Limit:      1,
				Query:      "Dillard",
				OrderField: OrderFieldId,
				OrderBy:    OrderByDesc,
			},
			expectedResponse: &SearchResponse{
				Users: []User{
					{
						Id:     17,
						Name:   "Dillard Mccoy",
						Age:    36,
						About:  "Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.",
						Gender: "male",
					},
				},
				NextPage: true,
			},
		},
		{
			request: SearchRequest{
				Limit:      1,
				Query:      "Dillard",
				OrderField: OrderFieldAge,
				OrderBy:    OrderByAsc,
			},
			expectedResponse: &SearchResponse{
				Users: []User{
					{
						Id:     3,
						Name:   "Everett Dillard",
						Age:    27,
						About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
		},
		{
			request: SearchRequest{
				OrderField: "LastName",
			},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("OrderField LastName invalid"),
		},
		{
			request: SearchRequest{
				OrderBy: 10,
			},
			expectedResponse: nil,
			expectedError:    fmt.Errorf(ErrorBadOrderBy),
		},
	}

	var testEnv = InitTestEnv(accessTokenCorrect)
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserLimits(t *testing.T) {
	var testCases = []TestCase{
		{
			request: SearchRequest{
				Limit: -1,
			},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("limit must be > 0"),
		},
		{
			request: SearchRequest{
				Limit:  100,
				Offset: 0,
			},
			expectedResponse: &SearchResponse{
				Users: make([]User, 25),
			},
			expectedError: nil,
		},
	}

	var testEnv = InitTestEnv(accessTokenCorrect)
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserOffsets(t *testing.T) {
	var testCases = []TestCase{
		{
			request: SearchRequest{
				Offset: -1,
			},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("offset must be > 0"),
		},
	}

	var testEnv = InitTestEnv(accessTokenCorrect)
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserWrongAccessToken(t *testing.T) {
	var testCases = []TestCase{
		{
			request: SearchRequest{
				Offset: 0,
			},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("bad AccessToken"),
		},
	}

	var testEnv = InitTestEnv("wrongAccessToken")
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserWrongNewRequest(t *testing.T) {
	var testCases = []TestCase{
		{
			request:          SearchRequest{},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("parse"),
		},
	}

	var testEnv = InitTestEnvWithUrl(accessTokenCorrect, "\n")
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserTimeout(t *testing.T) {
	var testCases = []TestCase{
		{
			request:          SearchRequest{},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("timeout"),
		},
	}

	var testEnv = InitTestEnvWithHandler(
		accessTokenCorrect,
		func(_ http.ResponseWriter, _ *http.Request) {
			time.Sleep(2 * time.Second)
		},
	)
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserUnknownError(t *testing.T) {
	var testCases = []TestCase{
		{
			request:          SearchRequest{},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("unknown error"),
		},
	}

	var testEnv = InitTestEnvWithUrl(accessTokenCorrect, "http://wrong-server")
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserCantUnpackResultJson(t *testing.T) {
	var testCases = []TestCase{
		{
			request:          SearchRequest{},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("cant unpack result json"),
		},
	}

	var testEnv = InitTestEnvWithHandler(
		accessTokenCorrect,
		func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, "Hello world")
		},
	)
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserCantUnpackErrorJson(t *testing.T) {
	var testCases = []TestCase{
		{
			request:          SearchRequest{},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("cant unpack error json"),
		},
	}

	var testEnv = InitTestEnvWithHandler(
		accessTokenCorrect,
		func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Hello world", http.StatusBadRequest)
		},
	)
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}

func TestFindUserFatalError(t *testing.T) {
	var testCases = []TestCase{
		{
			request:          SearchRequest{},
			expectedResponse: nil,
			expectedError:    fmt.Errorf("fatal error"),
		},
	}

	var testEnv = InitTestEnvWithHandler(
		accessTokenCorrect,
		func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "some error", http.StatusInternalServerError)
		},
	)
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)
		test.validate(t, response, err)
	}
}
