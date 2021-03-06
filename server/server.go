package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/GeertJohan/go.rice"

	"github.com/chancez/goplayutils/gist"
	"github.com/chancez/goplayutils/playground"
	"github.com/google/go-github/github"
)

var (
	templateBox *rice.Box
	templates   map[string]*template.Template
)

func init() {
	templates = make(map[string]*template.Template)
}

type Server struct {
	Client *github.Client
}

func getTemplate(name string) *template.Template {
	if templateBox == nil {
		templateBox = rice.MustFindBox("templates")
	}
	templateString := templateBox.MustString(name)
	if tmpl, ok := templates[name]; ok {
		return tmpl
	} else {
		templates[name] = template.Must(template.New(name).Parse(templateString))
		return templates[name]
	}
}

func indexHandler(rw http.ResponseWriter, req *http.Request) {
	index := getTemplate("index.html")
	index.Execute(rw, "")
}

func (server *Server) gistHandler(rw http.ResponseWriter, req *http.Request) {
	gistid := req.FormValue("gistid")
	if gistid != "" {
		log.Println("Received request for gistid:", gistid)
		content, err := gist.GetGist(server.Client, gistid)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			return
		}
		url, err := playground.GetPlayUrl(&content)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			return
		}
		res := map[string]string{"url": url}
		ct := req.Header.Get("Content-Type")

		if strings.Contains(ct, "json") {
			err := json.NewEncoder(rw).Encode(res)
			if err != nil {
				http.Error(rw, err.Error(), 500)
			}
			return
		}

		index := getTemplate("index.html")
		index.Execute(rw, res)
	} else {
		http.Redirect(rw, req, "/", http.StatusFound)
		return
	}
}

func (server *Server) registerHandlers() {
	// http.Handle("/", http.FileServer(rice.MustFindBox("templates").HTTPBox()))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/gist", server.gistHandler)
}

func (server *Server) Start(addr string) {
	server.registerHandlers()
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
