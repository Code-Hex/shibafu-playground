package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"cloud.google.com/go/datastore"
)

const hostname = "play.shibafu.dev"

var editTemplate = template.Must(template.ParseFiles("edit.html"))

type editData struct {
	Snippet   *snippet
	Analytics bool
}

func (s *server) editHandler() http.HandlerFunc {
	hostname := s.hostname
	return func(w http.ResponseWriter, r *http.Request) {
		// Redirect foo.play.golang.org to play.golang.org.
		if strings.HasSuffix(r.Host, "."+hostname) {
			http.Redirect(w, r, "https://"+hostname, http.StatusFound)
			return
		}

		// Serve 404 for /foo.
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/w/") {
			http.NotFound(w, r)
			return
		}

		snip := &snippet{Body: []byte(hello)}
		if strings.HasPrefix(r.URL.Path, "/w/") {
			id := r.URL.Path[3:]
			serveText := false
			if strings.HasSuffix(id, ".go") {
				id = id[:len(id)-3]
				serveText = true
			}

			if err := s.db.GetSnippet(r.Context(), id, snip); err != nil {
				if err != datastore.ErrNoSuchEntity {
					s.log.Errorf("loading Snippet: %v", err)
				}
				http.Error(w, "Snippet not found", http.StatusNotFound)
				return
			}
			if serveText {
				if r.FormValue("download") == "true" {
					w.Header().Set(
						"Content-Disposition", fmt.Sprintf(`attachment; filename="%s.w"`, id),
					)
				}
				w.Header().Set("Content-type", "text/plain; charset=utf-8")
				w.Write(snip.Body)
				return
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := &editData{
			Snippet:   snip,
			Analytics: r.Host == hostname,
		}
		if err := editTemplate.Execute(w, data); err != nil {
			s.log.Errorf("editTemplate.Execute(w, %+v): %v", data, err)
			return
		}
	}
}

const hello = `
  　    ∧,,∧
　　　 (；｀・ω・）　　,
　　　 /　ｏ={=}ｏ , ', ´
,,,,,しー-Ｊミ(.@)wwwwWwwWwwWwwWwwWwwWwwWwwWwwWw
wWWWWWwWwwWwwWwwWwwWwwWwwWwwWwwwwWwWWWw
WWWWwwwwwwWwwWwwWwwWwwWwwWwwWw
wWWWWWwWwwWwwWwwWwwwwWwWWWw
WWWwWwWwwwWwwWwwWwwWwwWwwWwwWwWwwWwwwWwwWwwWwWww
wWWWwWWWwwwwwWwwWwwWwwWwwWwwWwwWwwWw
wWWWWWwWwwWwwWwwWwwwwWwWWWwWWWWwwwwwwWwwWwwWwwWwwWwwWwwWwwWwwWwwWwwWw
wWWWWWwWwwWwwWwwWwwWwwwwWwWWWwWWWWwwwwwwWwwWwwWwwWwwWwwWwwWwwWw
wWWWWWwWwwWwwWwwwwWwWWWwWWWWwwwWwwWwwWwWwwWwWWwWWwWWwWWwWWwW
WwwWwWWwWWwWWwWWwWWwWWwWWwWWwwwWWWwWWWwwww
wWwwWwwWwwWwwWwwWwwWwwWw
wWWWWWwWwwWwwWwwWwwwwWwWWWw
WWWwWwWww
wWWWwWWWwwWwwWwwWwwWwwWwwWwwWwwWwwWwwWwWww
`
