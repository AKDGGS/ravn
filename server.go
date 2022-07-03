package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"

	"github.com/blevesearch/bleve/v2"
)

type WebServer struct {
	TaxonIndex    bleve.Index
	SpeciesIndex  bleve.Index
	GeneraIndex   bleve.Index
	ListenAddress string
	http          http.Server
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
	case "genera.json":
		uq := r.URL.Query()
		qv := uq.Get("q")
		sreq := bleve.NewSearchRequest(bleve.NewQueryStringQuery(qv))
		sreq.Fields = []string{"*"}
		sres, err := srv.GeneraIndex.Search(sreq)
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

	case "taxon.json":
		uq := r.URL.Query()
		qv := uq.Get("q")
		sreq := bleve.NewSearchRequest(bleve.NewQueryStringQuery(qv))
		sreq.Fields = []string{"*"}
		sres, err := srv.TaxonIndex.Search(sreq)
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
