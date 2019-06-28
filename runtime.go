package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Code-Hex/shibafu/evaluator"
)

type response struct {
	Errors string
	Events []Event
	Status int
}

func (s *server) runtimeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			// This is likely a pre-flight CORS request.
			return
		}

		body := r.FormValue("body")
		version := r.FormValue("version")

		s.log.Printf("version: %s, src: `%s`", version, body)

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		rec := new(Recorder)
		err := evaluator.New(body, os.Stdin, rec.Stdout()).Evaluate(ctx)
		if err != nil {
			var resp response
			if err == context.DeadlineExceeded {
				resp = response{Errors: "process took too long", Events: nil}
			} else {
				resp = response{Errors: err.Error(), Events: nil}
			}
			resp.Status = 1
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				s.log.Errorf("error encoding response: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}

		events, err := rec.Events()
		if err != nil {
			s.log.Errorf("recorder events error: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		resp := response{Events: events}
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(resp); err != nil {
			s.log.Errorf("error encoding response: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(w, &buf); err != nil {
			s.log.Errorf("io.Copy(w, &buf): %v", err)
			return
		}

	}
}
