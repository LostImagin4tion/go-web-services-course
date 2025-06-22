package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	usersDatabasePath = "./dataset.xml"

	userToken          = "row"
	userIdToken        = "id"
	userFirstNameToken = "first_name"
	userLastNameToken  = "last_name"
	userAgeToken       = "age"
	userAboutToken     = "about"
	userGenderToken    = "gender"
)

type SearchServer struct {
	server  http.Server
	usersDb []*User
}

func (s *SearchServer) StartServer(address string) {
	s.loadDb()

	var mux = http.NewServeMux()
	mux.HandleFunc("/", handler)

	s.server = http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	var err = s.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func (s *SearchServer) loadDb() {
	var file, err = os.Open(usersDatabasePath)
	if err != nil {
		panic(err)
	}

	var userId int
	var userNameBuilder strings.Builder
	var userAge int
	var userAbout string
	var userGender string

	var decoder = xml.NewDecoder(file)
	for {
		var token, tokenErr = decoder.Token()
		if tokenErr != nil && tokenErr != io.EOF {
			fmt.Println("Error happened", tokenErr)
			break
		} else if tokenErr == io.EOF {
			break
		}

		if token == nil {
			fmt.Println("Token is nil")
		}

		switch tokenType := token.(type) {
		case xml.StartElement:
			var decodeError error

			switch tokenType.Name.Local {
			case userIdToken:
				decodeError = decoder.DecodeElement(&userId, &tokenType)
			case userFirstNameToken:
				var userFirstName string
				decodeError = decoder.DecodeElement(&userFirstName, &tokenType)
				userNameBuilder.WriteString(userFirstName)
			case userLastNameToken:
				var userLastName string
				decodeError = decoder.DecodeElement(&userLastName, &tokenType)
				userNameBuilder.WriteString(userLastName)
			case userAgeToken:
				decodeError = decoder.DecodeElement(&userAge, &tokenType)
			case userAboutToken:
				decodeError = decoder.DecodeElement(&userAbout, &tokenType)
			case userGenderToken:
				decodeError = decoder.DecodeElement(&userGender, &tokenType)
			}

			if decodeError != nil {
				fmt.Printf("Error happened while decoding element %v, %e\n", tokenType, decodeError)
			}

		case xml.EndElement:
			switch tokenType.Name.Local {
			case userToken:
				var user = &User{
					Id:     userId,
					Name:   userNameBuilder.String(),
					Age:    userAge,
					About:  userAbout,
					Gender: userGender,
				}
				userNameBuilder.Reset()
				s.usersDb = append(s.usersDb, user)
			}
		}
	}

	fmt.Printf("Decoded %v users: %v", len(s.usersDb), s.usersDb)
}

func handler(w http.ResponseWriter, r *http.Request) {

}
