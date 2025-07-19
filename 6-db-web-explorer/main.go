package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"stepikGoWebServices/explorer"
	"stepikGoWebServices/queries"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// DSN это соединение с базой
	// вы можете изменить этот на тот который вам нужен
	// docker run -p 3306:3306 -v $(PWD):/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang -d mysql
	// DSN = "root@tcp(localhost:3306)/golang2017?charset=utf8"
	DSN = "root:1234@tcp(localhost:3306)/golang?charset=utf8&interpolateParams=true"
)

func main() {
	db, err := sql.Open("mysql", DSN)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	handler, err := explorer.NewDbExplorer(db, queries.NewQueriesMap())
	if err != nil {
		panic(err)
	}

	fmt.Println("starting server at :8082")
	http.ListenAndServe(":8082", &handler)
}
