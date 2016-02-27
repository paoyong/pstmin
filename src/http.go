package main

import (
    "fmt"
    "log"
    "html/template"

    "github.com/jackc/pgx"
    "github.com/buaazp/fasthttprouter"
    "github.com/valyala/fasthttp"
)

var (
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

    router := fasthttprouter.New()
    router.GET("/", Index)
    router.GET("/:paste_id", GrabPaste)
    router.POST("/save", Save)

    if db, err = initDatabase("localhost", "postgres", "postgres", "pastemin", 5432, 4); err != nil {
        log.Fatalf("Error opening database: %s", err)
    }

    log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
    fmt.Println("Listening on localhost:8080")
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
