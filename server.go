package main

import (
	"html/template"
	"log"
	"net/http"
)

var cmdServer = &Command{
	Run:       runServer,
	UsageLine: "server",
	Short:     "server",
}

var (
	serverWebDir string // -web flag
)

func init() {
	cmdServer.Flag.StringVar(&serverWebDir, "web", "", "")
}

func runServer(cmd *Command, args []string) {
	http.HandleFunc("/balls/", ballsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ballsHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[7:]
	p := &Page{title}
	t, _ := template.ParseFiles("layout.html")
	t.Execute(w, p)
}

type Page struct {
	Title string
}
