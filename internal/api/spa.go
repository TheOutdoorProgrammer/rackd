package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/TheOutdoorProgrammer/boating-accident/web"
)

// spaHandler serves the embedded React build, falling back to index.html so
// client-side routing works. Unknown /api/* paths get a JSON 404.
func (s *Server) spaHandler() http.HandlerFunc {
	dist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		panic(err) // embed is compiled in; a failure here is a build bug
	}
	fileServer := http.FileServer(http.FS(dist))

	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(dist, p); err != nil {
			// Unknown path: hand it to the SPA router via index.html.
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	}
}
