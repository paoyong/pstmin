package main

import (
    "fmt"
    "log"
    "html/template"

    "github.com/buaazp/fasthttprouter"
    "github.com/valyala/fasthttp"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func Index(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
    ctx.SetContentType("text/html")
    data := make([]string, 0, 10)
    tmpl.Execute(ctx, data)
}

func Save(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {
    fmt.Println(string(ctx.FormValue("pastearea")))
}

func main() {
    router := fasthttprouter.New()
    router.GET("/", Index)
    router.POST("/save", Save)

    log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}
