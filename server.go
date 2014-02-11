package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	d, _ := filepath.Abs(serverWebDir)
	serverWebDir = d

	http.HandleFunc("/", webHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/_js", jsHandler)
	http.HandleFunc("/_css", cssHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join(serverWebDir, "html/*.gohtml")
	log.Println(tmplPath)
	j, err := template.ParseGlob(tmplPath)
	if err != nil {
		panic(err)
	}

	// title := r.URL.Path[1:]
	p := &Page{"test"}
	err = j.ExecuteTemplate(w, "app.gohtml", p)
	if err != nil {
		panic(err)
	}

}

func webHandler(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join(serverWebDir, "html/*.gohtml")
	log.Println(tmplPath)
	j, err := template.ParseGlob(tmplPath)
	if err != nil {
		panic(err)
	}

	// title := r.URL.Path[1:]
	p := &Page{"test"}
	err = j.ExecuteTemplate(w, "index.gohtml", p)
	if err != nil {
		panic(err)
	}
	// fmt.Fprintf(w, "FML")
	// log.Println(j.Lookup("index.html"))

	// j.Lookup("index.html").Execute(w, p)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join(serverWebDir, "html/*.gohtml")
	log.Println(tmplPath)
	j, err := template.ParseGlob(tmplPath)
	if err != nil {
		panic(err)
	}
	p := &Page{"test"}
	err = j.ExecuteTemplate(w, "login.gohtml", p)
	if err != nil {
		panic(err)
	}
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("js")
	w.Header().Add("Content-Type", "application/javascript")

	files, err := filepath.Glob(filepath.Join(serverWebDir, "/js/*.js"))
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
	// fmt.Fprintf(w, `alert("lol");`)
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("css")
	w.Header().Add("Content-Type", "text/css")
	files, err := filepath.Glob(filepath.Join(serverWebDir, "/css/*.css"))
	if err != nil {
		panic(err)
	}

	for _, filename := range files {
		log.Println(filename)
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

type Page struct {
	Title string
}

// type Stylesheet struct {

// }
