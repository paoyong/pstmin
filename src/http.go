package main

import (
    "fmt"
    "html/template"
    "log"
    "math/rand"
    "time"
    "io/ioutil"
    "encoding/json"
    "strconv"

    "github.com/buaazp/fasthttprouter"
    "github.com/jackc/pgx"
    "github.com/valyala/fasthttp"
)

type Configuration struct {
    DBHost      string `json:"db_host"`
    DBUser      string  `json:"db_user"`
    DBPass      string  `json:"db_pass"`
    DBName      string  `json:"db_name"`
    DBPort      string  `json:"db_port"`
}

var (
    grabPasteById *pgx.PreparedStatement
    insertPaste *pgx.PreparedStatement
    tmpl = template.Must(template.ParseFiles("templates/index.html"))
    config = grabConfig("config.json")
    db *pgx.ConnPool
    idAlphabet = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    idNumChars = 8
)

func main() {
    init()

    router := fasthttprouter.New()
    router.GET("/", Index)
    router.GET("/:paste_id", GrabPaste)
    router.POST("/save", Save)

    log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}

func init() {
    rand.Seed(time.Now().UnixNano())

    port, err := strconv.ParseUint(config.DBPort, 10, 16)
    if db, err = initDatabase(config.DBHost, config.DBUser, config.DBPass, config.DBName, uint16(port), 256); err != nil {
        log.Fatalf("Error opening database: %s", err)
    }
}

func grabConfig(filename string) *Configuration {
    c := Configuration{}

    b, err := ioutil.ReadFile(filename)
    if err != nil {
        log.Fatalf("Error reading in config file: %s", err)
    }

    if err := json.Unmarshal(b, &c); err != nil {
        log.Fatalf("Error decoding config file: %s", err)
    }

    return &c
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
    var pasteText string
    pasteId := ps.ByName("paste_id")

    if err := db.QueryRow("grabPasteById", pasteId).Scan(&pasteText); err != nil {
        if err.Error() == "no rows in result set" {
            ctx.SetStatusCode(fasthttp.StatusNotFound)
            fmt.Fprintf(ctx, "Can't find that paste!")
        } else {
            log.Fatalf("Error grabbing paste: %s", err)
        }
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

    config.AfterConnect = func(conn *pgx.Conn) error {
        var err error

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
        fmt.Println("Connecting to database", dbName, "as user", dbUser, ":", successOrFailure, err)
    } else {
        fmt.Println("Connecting to database", dbName, "as user", dbUser, ":", successOrFailure)
    }

    return connPool, err
}
