package main

import (
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"github.com/crowdmob/goamz/s3"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"willstclair.com/phosphorus/app"
	"willstclair.com/phosphorus/id"
)

var cmdServer = &Command{
	Run:       runServer,
	UsageLine: "server",
	Short:     "server",
}

var (
	serverWebDir  string // -web flag
	serverUrlRoot string // -url flag

)

var t = make(map[string]*template.Template)

func init() {
	cmdServer.Flag.StringVar(&serverWebDir, "web", "", "")
	cmdServer.Flag.StringVar(&serverUrlRoot, "url", "", "")
}

func initTemplates() {
	joined := filepath.Join(serverWebDir, "html/*.html")
	tmplDir, err := filepath.Abs(joined)
	if err != nil {
		panic(err)
	}

	templates, err := filepath.Glob(tmplDir)
	if err != nil {
		panic(err)
	}

	for _, filename := range templates {
		bn := filepath.Base(filename)
		if bn == "layout.html" {
			continue
		}
		t[bn] = template.Must(template.ParseFiles(filename,
			filepath.Join(serverWebDir, "html", "layout.html")))
	}

}

func runServer(cmd *Command, args []string) {
	initTemplates()
	log.Println(t)

	exp := time.Now().Add(60 * time.Minute)
	auth, err := aws.EnvAuth()
	if err != nil {
		auth, err = aws.GetAuth("", "", "", exp)
		if err != nil {
			panic(err)
		}
	}

	dynamo := &dynamodb.Server{auth, aws.USEast}
	sessionTbl := dynTable(dynamo, "session")
	accountTbl := dynTable(dynamo, "account")

	s3server := s3.New(auth, aws.USEast)
	bucket := &s3.Bucket{s3server, "phosphorus-upload"}

	idGen := id.NewGenerator(1)

	http.HandleFunc("/enroll", app.CreateAccountHandler(t, accountTbl, sessionTbl, idGen))
	http.HandleFunc("/login", app.LoginHandler(t, accountTbl, sessionTbl, idGen))
	http.HandleFunc("/u/", app.DashboardHandler(t, accountTbl, sessionTbl, idGen, bucket))

	http.HandleFunc("/_form", app.UploadTemplateHandler(accountTbl, sessionTbl, idGen))

	http.HandleFunc("/_js", jsHandler)
	http.HandleFunc("/_css", cssHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func catGlob(ext string, w http.ResponseWriter) {
	files, err := filepath.Glob(
		filepath.Join(
			serverWebDir,
			fmt.Sprintf("/%s/*.%s", ext, ext)))
	if err != nil {
		panic(err)
	}

	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(w, file)
		if err != nil {
			panic(err)
		}
		file.Close()
	}
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/javascript")
	catGlob("js", w)
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/css")
	catGlob("css", w)
}

func render(w http.ResponseWriter, tmpl string) {
	err := t[tmpl+".html"].ExecuteTemplate(w, "layout", nil)
	if err != nil {
		panic(err)
	}
}
