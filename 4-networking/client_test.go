package main

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

type TestCase struct {
	request          SearchRequest
	expectedResponse SearchResponse
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

func TestFindUser(t *testing.T) {
	var testCases = []TestCase{
		{
			request: SearchRequest{
				Limit: 1,
				Query: "Everett",
			},
			expectedResponse: SearchResponse{
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
	}

	var testEnv = InitTestEnv(accessTokenCorrect)
	var client = testEnv.Client
	defer testEnv.Server.Close()

	for _, test := range testCases {
		var response, err = client.FindUsers(test.request)

		if err != nil {
			t.Errorf("Some error happened: %v", err)
		}
		if len(response.Users) != len(test.expectedResponse.Users) {
			t.Errorf(
				"Invalid number of users. Got %v, expected %v",
				len(response.Users),
				len(test.expectedResponse.Users),
			)
		}
		if !slices.Equal(response.Users, test.expectedResponse.Users) {
			t.Errorf(
				"Got wrong response. Got %v\nExpected %v",
				response.Users,
				test.expectedResponse.Users,
			)
		}
	}
}
