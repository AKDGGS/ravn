package assets

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

type staticEntry struct {
	ModTime   time.Time
	Content   *[]byte
	ETag      string
	GZContent *[]byte
	GZETag    string
}

var staticLock sync.RWMutex
var staticCache map[string]*staticEntry = make(map[string]*staticEntry)

func ServeStatic(name string, w http.ResponseWriter, r *http.Request) {
	s, err := Stat(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(
				w, fmt.Sprintf("stat error: %s", err.Error()),
				http.StatusInternalServerError,
			)
		}
		return
	}

	staticLock.RLock()
	curEntry, ok := staticCache[name]
	staticLock.RUnlock()

	if !ok || s.ModTime().After(curEntry.ModTime) {
		b, err := ReadBytes(name)
		if err != nil {
			http.Error(
				w, fmt.Sprintf("read error: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}

		curEntry = &staticEntry{
			ModTime: s.ModTime(),
			Content: &b,
			ETag:    fmt.Sprintf("%x", md5.Sum(b)),
		}

		// 860 bytes is Akamai's recommended minimum for gzip
		// so only bother to gzip files greater than 860 bytes
		if len(b) > 860 {
			var buf bytes.Buffer
			gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
			if err != nil {
				http.Error(
					w, fmt.Sprintf("gzip error: %s", err.Error()),
					http.StatusInternalServerError,
				)
				return
			}
			defer gz.Close()

			if _, err := gz.Write(b); err != nil {
				http.Error(
					w, fmt.Sprintf("gz write error: %s", err.Error()),
					http.StatusInternalServerError,
				)
				return
			}

			if err := gz.Flush(); err != nil {
				http.Error(
					w, fmt.Sprintf("gz write error: %s", err.Error()),
					http.StatusInternalServerError,
				)
				return
			}

			// Only accept gzip if it's less than the original in size
			if buf.Len() > 0 && buf.Len() < len(b) {
				gzc := buf.Bytes()
				curEntry.GZContent = &gzc
				curEntry.GZETag = fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
			}
		}
		staticLock.Lock()
		staticCache[name] = curEntry
		staticLock.Unlock()
	}

	var content *[]byte
	var etag *string
	gzok := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && curEntry.GZContent != nil
	if gzok {
		content = curEntry.GZContent
		etag = &curEntry.GZETag
	} else {
		content = curEntry.Content
		etag = &curEntry.ETag
	}

	retag := r.Header.Get("If-None-Match")
	if *etag == retag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", *etag)
	contenttype := mime.TypeByExtension(path.Ext(name))
	if contenttype == "" {
		contenttype = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contenttype)

	if gzok {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(*content)))
	w.Write(*content)
}
