package main

import (
	"database/sql"
	"fmt"
	"io"
	"reflect"
	"stepikGoWebServices/explorer"
	"stepikGoWebServices/queries"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type CaseResponse map[string]any

type Case struct {
	Method string
	Path   string
	Query  string
	Status int
	Result any
	Body   any
}

var (
	client = &http.Client{Timeout: time.Second}
)

func PrepareTestApis(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS items;`,

		`CREATE TABLE items (
  id int(11) NOT NULL AUTO_INCREMENT,
  title varchar(255) NOT NULL,
  description text NOT NULL,
  updated varchar(255) DEFAULT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,

		`INSERT INTO items (id, title, description, updated) VALUES
(1,	'database/sql',	'Рассказать про базы данных',	'rvasily'),
(2,	'memcache',	'Рассказать про мемкеш с примером использования',	NULL);`,

		`DROP TABLE IF EXISTS users;`,

		`CREATE TABLE users (
			user_id int(11) NOT NULL AUTO_INCREMENT,
  login varchar(255) NOT NULL,
  password varchar(255) NOT NULL,
  email varchar(255) NOT NULL,
  info text NOT NULL,
  updated varchar(255) DEFAULT NULL,
  PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,

		`INSERT INTO users (user_id, login, password, email, info, updated) VALUES
(1,	'rvasily',	'love',	'rvasily@example.com',	'none',	NULL);`,
	}

	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func CleanupTestApis(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS items;`,
		`DROP TABLE IF EXISTS users;`,
	}
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func TestApis(t *testing.T) {
	db, err := sql.Open("mysql", DSN)
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	PrepareTestApis(db)

	defer CleanupTestApis(db)

	handler, err := explorer.NewDbExplorer(db, queries.NewQueriesMap())
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(&handler)

	cases := []Case{
		{
			Path: "/", // список таблиц
			Result: CaseResponse{
				"response": CaseResponse{
					"tables": []string{"items", "users"},
				},
			},
		},
		{
			Path:   "/unknown_table",
			Status: http.StatusNotFound,
			Result: CaseResponse{
				"error": "unknown table",
			},
		},
		{
			Path: "/items",
			Result: CaseResponse{
				"response": CaseResponse{
					"records": []CaseResponse{
						{
							"id":          1,
							"title":       "database/sql",
							"description": "Рассказать про базы данных",
							"updated":     "rvasily",
						},
						{
							"id":          2,
							"title":       "memcache",
							"description": "Рассказать про мемкеш с примером использования",
							"updated":     nil,
						},
					},
				},
			},
		},
		{
			Path:  "/items",
			Query: "limit=1",
			Result: CaseResponse{
				"response": CaseResponse{
					"records": []CaseResponse{
						{
							"id":          1,
							"title":       "database/sql",
							"description": "Рассказать про базы данных",
							"updated":     "rvasily",
						},
					},
				},
			},
		},
		{
			Path:  "/items",
			Query: "limit=1&offset=1",
			Result: CaseResponse{
				"response": CaseResponse{
					"records": []CaseResponse{
						{
							"id":          2,
							"title":       "memcache",
							"description": "Рассказать про мемкеш с примером использования",
							"updated":     nil,
						},
					},
				},
			},
		},
		{
			Path: "/items/1",
			Result: CaseResponse{
				"response": CaseResponse{
					"record": CaseResponse{
						"id":          1,
						"title":       "database/sql",
						"description": "Рассказать про базы данных",
						"updated":     "rvasily",
					},
				},
			},
		},
		{
			Path:   "/items/100500",
			Status: http.StatusNotFound,
			Result: CaseResponse{
				"error": "record not found",
			},
		},

		// тут идёт создание и редактирование
		{
			Path:   "/items/",
			Method: http.MethodPut,
			Body: CaseResponse{
				"id":          42, // auto increment primary key игнорируется при вставке
				"title":       "db_crud",
				"description": "",
			},
			Result: CaseResponse{
				"response": CaseResponse{
					"id": 3,
				},
			},
		},
		// это пример хрупкого теста
		// если много раз вызывать один и тот же тест - записи будут добавляться
		// поэтому придётся сделать сброс базы каждый раз в PrepareTestData
		{
			Path: "/items/3",
			Result: CaseResponse{
				"response": CaseResponse{
					"record": CaseResponse{
						"id":          3,
						"title":       "db_crud",
						"description": "",
						"updated":     nil,
					},
				},
			},
		},
		{
			Path:   "/items/3",
			Method: http.MethodPost,
			Body: CaseResponse{
				"description": "Написать программу db_crud",
			},
			Result: CaseResponse{
				"response": CaseResponse{
					"updated": 1,
				},
			},
		},
		{
			Path: "/items/3",
			Result: CaseResponse{
				"response": CaseResponse{
					"record": CaseResponse{
						"id":          3,
						"title":       "db_crud",
						"description": "Написать программу db_crud",
						"updated":     nil,
					},
				},
			},
		},

		// обновление null-поля в таблице
		{
			Path:   "/items/3",
			Method: http.MethodPost,
			Body: CaseResponse{
				"updated": "autotests",
			},
			Result: CaseResponse{
				"response": CaseResponse{
					"updated": 1,
				},
			},
		},
		{
			Path: "/items/3",
			Result: CaseResponse{
				"response": CaseResponse{
					"record": CaseResponse{
						"id":          3,
						"title":       "db_crud",
						"description": "Написать программу db_crud",
						"updated":     "autotests",
					},
				},
			},
		},

		// обновление null-поля в таблице
		{
			Path:   "/items/3",
			Method: http.MethodPost,
			Body: CaseResponse{
				"updated": nil,
			},
			Result: CaseResponse{
				"response": CaseResponse{
					"updated": 1,
				},
			},
		},
		{
			Path: "/items/3",
			Result: CaseResponse{
				"response": CaseResponse{
					"record": CaseResponse{
						"id":          3,
						"title":       "db_crud",
						"description": "Написать программу db_crud",
						"updated":     nil,
					},
				},
			},
		},

		// ошибки
		{
			Path:   "/items/3",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Body: CaseResponse{
				"id": 4, // primary key нельзя обновлять у существующей записи
			},
			Result: CaseResponse{
				"error": "field id have invalid type",
			},
		},
		{
			Path:   "/items/3",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Body: CaseResponse{
				"title": 42,
			},
			Result: CaseResponse{
				"error": "field title have invalid type",
			},
		},
		{
			Path:   "/items/3",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Body: CaseResponse{
				"title": nil,
			},
			Result: CaseResponse{
				"error": "field title have invalid type",
			},
		},

		{
			Path:   "/items/3",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Body: CaseResponse{
				"updated": 42,
			},
			Result: CaseResponse{
				"error": "field updated have invalid type",
			},
		},

		// удаление
		{
			Path:   "/items/3",
			Method: http.MethodDelete,
			Result: CaseResponse{
				"response": CaseResponse{
					"deleted": 1,
				},
			},
		},
		{
			Path:   "/items/3",
			Method: http.MethodDelete,
			Result: CaseResponse{
				"response": CaseResponse{
					"deleted": 0,
				},
			},
		},
		{
			Path:   "/items/3",
			Status: http.StatusNotFound,
			Result: CaseResponse{
				"error": "record not found",
			},
		},

		// и немного по другой таблице
		{
			Path: "/users/1",
			Result: CaseResponse{
				"response": CaseResponse{
					"record": CaseResponse{
						"user_id":  1,
						"login":    "rvasily",
						"password": "love",
						"email":    "rvasily@example.com",
						"info":     "none",
						"updated":  nil,
					},
				},
			},
		},

		{
			Path:   "/users/1",
			Method: http.MethodPost,
			Body: CaseResponse{
				"info":    "try update",
				"updated": "now",
			},
			Result: CaseResponse{
				"response": CaseResponse{
					"updated": 1,
				},
			},
		},
		{
			Path: "/users/1",
			Result: CaseResponse{
				"response": CaseResponse{
					"record": CaseResponse{
						"user_id":  1,
						"login":    "rvasily",
						"password": "love",
						"email":    "rvasily@example.com",
						"info":     "try update",
						"updated":  "now",
					},
				},
			},
		},
		// ошибки
		{
			Path:   "/users/1",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Body: CaseResponse{
				"user_id": 1, // primary key нельзя обновлять у существующей записи
			},
			Result: CaseResponse{
				"error": "field user_id have invalid type",
			},
		},
		// не забываем про sql-инъекции
		{
			Path:   "/users/",
			Method: http.MethodPut,
			Body: CaseResponse{
				"user_id":    2,
				"login":      "qwerty'",
				"password":   "love\"",
				"unkn_field": "love",
			},
			Result: CaseResponse{
				"response": CaseResponse{
					"user_id": 2,
				},
			},
		},
		{
			Path: "/users/2",
			Result: CaseResponse{
				"response": CaseResponse{
					"record": CaseResponse{
						"user_id":  2,
						"login":    "qwerty'",
						"password": "love\"",
						"email":    "",
						"info":     "",
						"updated":  nil,
					},
				},
			},
		},
		// тут тоже возможна sql-инъекция
		// если пришло не число на вход - берём дефолтное значение для лимита-офсета
		{
			Path:  "/users",
			Query: "limit=1'&offset=1\"",
			Result: CaseResponse{
				"response": CaseResponse{
					"records": []CaseResponse{
						{
							"user_id":  1,
							"login":    "rvasily",
							"password": "love",
							"email":    "rvasily@example.com",
							"info":     "try update",
							"updated":  "now",
						},
						{
							"user_id":  2,
							"login":    "qwerty'",
							"password": "love\"",
							"email":    "",
							"info":     "",
							"updated":  nil,
						},
					},
				},
			},
		},
	}

	runCases(t, ts, db, cases)
}

func runCases(t *testing.T, ts *httptest.Server, db *sql.DB, cases []Case) {
	for idx, item := range cases {
		var (
			err      error
			result   any
			expected any
			req      *http.Request
		)

		caseName := fmt.Sprintf("case %d: [%s] %s %s", idx, item.Method, item.Path, item.Query)

		// если у вас случилась это ошибка - значит вы не делаете где-то rows.Close и у вас текут соединения с базой
		// если такое случилось на первом тесте - значит вы не закрываете коннект где-то при инициализации в NewDbExplorer
		if db.Stats().OpenConnections != 1 {
			t.Fatalf("[%s] you have %d open connections, must be 1", caseName, db.Stats().OpenConnections)
		}

		if item.Method == "" || item.Method == http.MethodGet {
			req, err = http.NewRequest(item.Method, ts.URL+item.Path+"?"+item.Query, nil)
		} else {
			data, err := json.Marshal(item.Body)
			if err != nil {
				panic(err)
			}
			reqBody := bytes.NewReader(data)
			req, err = http.NewRequest(item.Method, ts.URL+item.Path, reqBody)
			req.Header.Add("Content-Type", "application/json")
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("[%s] request error: %v", caseName, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)

		if item.Status == 0 {
			item.Status = http.StatusOK
		}

		if resp.StatusCode != item.Status {
			t.Fatalf("[%s] expected http status %v, got %v", caseName, item.Status, resp.StatusCode)
			continue
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatalf("[%s] cant unpack json: %v", caseName, err)
			continue
		}

		// reflect.DeepEqual не работает если нам приходят разные типы
		// а там приходят разные типы (string VS interface{}) по сравнению с тем что в ожидаемом результате
		// этот маленький грязный хак конвертит данные сначала в json, а потом обратно в interface - получаем совместимые результаты
		// не используйте это в продакшен-коде - надо явно писать что ожидается интерфейс или использовать другой подход с точным форматом ответа
		data, err := json.Marshal(item.Result)
		json.Unmarshal(data, &expected)

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("[%s] results not match\nGot : %#v\nWant: %#v", caseName, result, expected)
			continue
		}
	}

}
