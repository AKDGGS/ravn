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
	ImagesPath      string
	KeyFile         string
	CertFile        string
	http            http.Server
}

func (srv *WebServer) Start() error {
	listen, err := net.Listen("tcp", srv.ListenAddress)
	if err != nil {
		return err
	}

	fmt.Printf("listening on %s .. \n", listen.Addr().String())

	srv.http = http.Server{Handler: srv}
	if srv.CertFile != "" && srv.KeyFile != "" {
		err = srv.http.ServeTLS(listen, srv.CertFile, srv.KeyFile)
	} else {
		err = srv.http.Serve(listen)
	}
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (srv *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch path.Base(r.URL.Path) {
	case "/", "index.html":
		assets.ServeStatic("html/index.html", w, r)
		return

	case "search.js":
		assets.ServeStatic("js/search.js", w, r)
		return

	case "template.css":
		assets.ServeStatic("css/template.css", w, r)
		return

	case "index.css":
		assets.ServeStatic("css/index.css", w, r)
		return

	case "genera_full.json":
		q := r.URL.Query()
		doc, err := indexDocID(srv.GeneraIndex, []string{"*"}, q.Get("id"))
		if err != nil {
			http.Error(
				w, fmt.Sprintf("document error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}

		if doc == nil {
			http.Error(w, "Document not found", http.StatusNotFound)
		}

		jb, err := json.Marshal(doc)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("json error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jb)
		return

	case "genera.json":
		q := r.URL.Query()
		sres, err := indexQuery(
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
		return

	case "references.json":
		q := r.URL.Query()
		sres, err := indexQuery(
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
		return

	case "species_full.json":
		q := r.URL.Query()
		doc, err := indexDocID(srv.SpeciesIndex, []string{"*"}, q.Get("id"))
		if err != nil {
			http.Error(
				w, fmt.Sprintf("document error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}

		if doc == nil {
			http.Error(w, "Document not found", http.StatusNotFound)
		}

		jb, err := json.Marshal(doc)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("json error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jb)
		return

	case "species.json":
		q := r.URL.Query()
		sres, err := indexQuery(
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
		return
	}

	if srv.ImagesPath != "" && path.Base(path.Dir(r.URL.Path)) == "images" {
		filename := path.Base(r.URL.Path)
		http.ServeFile(w, r, path.Join(srv.ImagesPath, filename))
		return
	}

	http.Error(w, "File not found", http.StatusNotFound)
	return
}

func indexDocID(idx bleve.Index, fields []string, id string) (map[string]interface{}, error) {
	sreq := bleve.NewSearchRequest(
		bleve.NewDocIDQuery([]string{id}),
	)
	sreq.Fields = fields

	sres, err := idx.Search(sreq)
	if err != nil {
		return nil, err
	}

	if len(sres.Hits) < 1 {
		return make(map[string]interface{}, 0), nil
	}
	ret := sres.Hits[0].Fields
	ret["time"] = sres.Took.String()
	ret["id"] = sres.Hits[0].ID
	return ret, nil
}

func indexQuery(idx bleve.Index, fields []string, q, sz, fr string) (map[string]interface{}, error) {
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
