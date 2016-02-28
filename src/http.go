package main

import (
    "fmt"
    "html/template"
    "log"
    "math/rand"
    "time"
    "io/ioutil"
    "encoding/json"

    "github.com/buaazp/fasthttprouter"
    "github.com/jackc/pgx"
    "github.com/valyala/fasthttp"
)

type Configuration struct {
    DBUser     string `json:"db_user"`
    DBPass     string `json:"db_pass"`
    DBName     string `json:"db_name"`
    DBPort     string `json:"db_port"`
}

var (
    grabPasteById *pgx.PreparedStatement
    insertPaste *pgx.PreparedStatement
    tmpl = template.Must(template.ParseFiles("templates/index.html"))
    db *pgx.ConnPool
    idAlphabet = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    idNumChars = 8
)


func main() {
    rand.Seed(time.Now().UnixNano())

    // Parse config file into config object
    config := Configuration{}
    b, err := ioutil.ReadFile("config.json")
    if err != nil {
        log.Fatalf("Error reading config.json: %s", err)
    }
    if err := json.Unmarshal(b, &config); err != nil {
        fmt.Println("Error decoding config file", err)
    }

    if db, err = initDatabase(config.DBHost, config.DBUser, config.DBPass, config.DBName, 5432, 256); err != nil {
        log.Fatalf("Error opening database: %s", err)
    }

    router := fasthttprouter.New()
    router.GET("/", Index)
    router.GET("/:paste_id", GrabPaste)
    router.POST("/save", Save)

    fmt.Println("Listening on localhost:8080")
    log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}

func Index(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
    ctx.SetContentType("text/html")
    data := make([]string, 0, 10)
    tmpl.Execute(ctx, data)
}

func Save(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {
    randId := generateRandomId(idNumChars, idAlphabet)
    paste := string(ctx.FormValue("pastearea"))

    txn, err := db.Begin()
    if err != nil {
        log.Fatalf("Error starting db: %s", err)
    }

    if _, err := txn.Exec("insertPaste", randId, paste); err != nil {
        log.Fatalf("Error inserting new paste: %s", err)
    }

    if err = txn.Commit(); err != nil {
        log.Fatalf("Error when committing new paste: %s", err)
    }

    ctx.Redirect("/" + randId, 302)
}

func GrabPaste(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {
    pasteId := ps.ByName("paste_id")

    var pasteText string

    if err := db.QueryRow("grabPasteById", pasteId).Scan(&pasteText); err != nil {
        log.Fatalf("Error grabbing paste: %s", err)
    }

    fmt.Fprint(ctx, pasteText)
}

/* generateRandomId
* ------------
* Returns an random alphanumeric string of length numChars, like xJ2h9a0 */
func generateRandomId(numChars int, alphabet []rune) string {
    randStr := make([]rune, numChars)
    for i := range randStr {
        randStr[i] = alphabet[rand.Intn(len(alphabet))]
    }

    return string(randStr)
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
        grabPasteById, err = conn.Prepare("grabPasteById", "SELECT paste FROM pastes WHERE id = $1")
        if err != nil {
            log.Fatalf("Error when preparing statement grabPaste: %s", err)
        }

        insertPaste, err = conn.Prepare("insertPaste", "INSERT INTO pastes(id, paste) VALUES($1, $2)")
        if err != nil {
            log.Fatalf("Error when preparing statement grabPasteById: %s", err)
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

    connPool, err := pgx.NewConnPool(config)
    if err != nil {
        successOrFailure = "FAILED"
        fmt.Println("Connecting to database ", dbName, " as user ", dbUser, " ", successOrFailure, ": \n ", err)
    } else {
        fmt.Println("Connecting to database ", dbName, " as user ", dbUser, ": ", successOrFailure)
    }

    return connPool, err
}
