package main

import (
	"fmt"
	"html/template"
	"log"

	"github.com/buaazp/fasthttprouter"
	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

var (
	grabPaste *pgx.PreparedStatement

	tmpl = template.Must(template.ParseFiles("templates/index.html"))

	db *pgx.ConnPool
)

func Index(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
	ctx.SetContentType("text/html")
	data := make([]string, 0, 10)
	tmpl.Execute(ctx, data)
}

func Save(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {
	fmt.Println(string(ctx.FormValue("pastearea")))
	ctx.Redirect("/", 302)
}

func GrabPaste(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {

}

func main() {
	var err error

	if db, err = initDatabase("localhost", "postgres", "postgres", "pastemin", 5432, 4); err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	router := fasthttprouter.New()
	router.GET("/", Index)
	router.GET("/:paste_id", GrabPaste)
	router.POST("/save", Save)

        rows, err := db.Query("grabPaste")

	fmt.Println("Listening on localhost:8080")
	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}

func generateRandomId(numChars uint16) {
    const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    // TODO
}

/* initDatabase
 * ------------
 * Initializes database connection.
 * Taken from https://github.com/TechEmpower/FrameworkBenchmarks/blob/master/frameworks/Go/fasthttp-postgresql/src/hello/hello.go */
func initDatabase(dbHost string, dbUser string, dbPass string, dbName string, dbPort uint16, maxConnectionsInPool int) (*pgx.ConnPool, error) {
	var successOrFailure string = "OK"

	var config pgx.ConnPoolConfig

	config.Host = dbHost
	config.User = dbUser
	config.Password = dbPass
	config.Database = dbName
	config.Port = dbPort

	config.MaxConnections = maxConnectionsInPool

        var err error
	config.AfterConnect = func(conn *pgx.Conn) error {
            grabPaste, err = conn.Prepare("grabPaste", "SELECT * FROM pastes WHERE id = $1")
            if err != nil {
                    log.Fatalf("Error when preparing statement grabPaste: %s", err)
            }

            // Disable synchronous commit for the current db connection
            // as a performance optimization.
            // See http://www.postgresql.org/docs/current/static/runtime-config-wal.html
            // for details.
            if _, err := conn.Exec("SET synchronous_commit TO OFF"); err != nil {
                    log.Fatalf("Error when disabling synchronous commit")
            }

            return nil
	}

	fmt.Println("--------------------------------------------------------------------------------------------")

	connPool, err := pgx.NewConnPool(config)
	if err != nil {
		successOrFailure = "FAILED"
		fmt.Println("Connecting to database ", dbName, " as user ", dbUser, " ", successOrFailure, ": \n ", err)
	} else {
		fmt.Println("Connecting to database ", dbName, " as user ", dbUser, ": ", successOrFailure)
	}

	fmt.Println("--------------------------------------------------------------------------------------------")

	return connPool, err
}
