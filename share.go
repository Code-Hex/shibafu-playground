package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

const (
	// This salt is not meant to be kept secret (it’s checked in after all). It’s
	// a tiny bit of paranoia to avoid whatever problems a collision may cause.
	salt = "Shibafu playground salt\n"

	maxSnippetSize = 128 * 1024
)

func (s *server) shareHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			// This is likely a pre-flight CORS request.
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Requires POST", http.StatusMethodNotAllowed)
			return
		}

		var body bytes.Buffer
		_, err := io.Copy(&body, io.LimitReader(r.Body, maxSnippetSize+1))
		r.Body.Close()
		if err != nil {
			s.log.Errorf("reading Body: %v", err)
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		if body.Len() > maxSnippetSize {
			http.Error(w, "Snippet is too large", http.StatusRequestEntityTooLarge)
			return
		}

		snip := &snippet{Body: body.Bytes()}
		id := snip.ID()
		if err := s.db.PutSnippet(r.Context(), id, snip); err != nil {
			s.log.Errorf("putting Snippet: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, id)
	}
}
