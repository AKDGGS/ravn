package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"

	"ravn/assets"

	"github.com/blevesearch/bleve/v2"
)

type WebServer struct {
	ReferencesIndex bleve.Index
	SpeciesIndex    bleve.Index
	GeneraIndex     bleve.Index
	ListenAddress   string
	http            http.Server
}

func (srv *WebServer) Start() error {
	listen, err := net.Listen("tcp", srv.ListenAddress)
	if err != nil {
		return err
	}

	fmt.Printf("listening on %s .. \n", listen.Addr().String())

	srv.http = http.Server{Handler: srv}
	err = srv.http.Serve(listen)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (srv *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch path.Base(r.URL.Path) {
	case "/", "index.html":
		assets.ServeStatic("html/index.html", w, r)

	case "search.js":
		assets.ServeStatic("js/search.js", w, r)

	case "genera.json":
		q := r.URL.Query()
		sres, err := searchIndex(
			srv.GeneraIndex, []string{"source", "alt_source"},
			q.Get("q"), q.Get("z"), q.Get("f"),
		)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("search error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}

		jb, err := json.Marshal(sres)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("json error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jb)

	case "references.json":
		q := r.URL.Query()
		sres, err := searchIndex(
			srv.ReferencesIndex, []string{"*"},
			q.Get("q"), q.Get("z"), q.Get("f"),
		)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("search error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}

		jb, err := json.Marshal(sres)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("json error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jb)

	case "species.json":
		q := r.URL.Query()
		sres, err := searchIndex(
			srv.SpeciesIndex, []string{"source", "alt_source"},
			q.Get("q"), q.Get("z"), q.Get("f"),
		)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("search error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}

		jb, err := json.Marshal(sres)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("json error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jb)

	default:
		http.Error(w, "File not found", http.StatusNotFound)
	}
}

func searchIndex(idx bleve.Index, fields []string, q, sz, fr string) (map[string]interface{}, error) {
	size, _ := strconv.Atoi(sz)
	if size < 1 {
		size = 25
	}
	from, _ := strconv.Atoi(fr)
	if from < 0 {
		from = 0
	}

	sreq := bleve.NewSearchRequest(
		bleve.NewQueryStringQuery(strings.ToLower(q)),
	)
	sreq.Fields = fields
	sreq.Size = size
	sreq.From = from
	sres, err := idx.Search(sreq)
	if err != nil {
		return nil, err
	}

	hits := make([]map[string]interface{}, 0)
	for _, hit := range sres.Hits {
		r := make(map[string]interface{})
		for k, v := range hit.Fields {
			r[k] = v
		}
		r["id"] = hit.ID
		hits = append(hits, r)
	}

	wrap := make(map[string]interface{}, 0)
	wrap["hits"] = hits
	wrap["total"] = sres.Total
	wrap["time"] = sres.Took.String()

	return wrap, nil
}
