package routers

import (
	"net/http"

	"github.com/mudkipme/lilycove/lib"
)

// Init web routers
func Init(s *http.ServeMux) {
	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PURGE" {
			purger := lib.DefaultPurger()
			purger.Add(r.Host, r.URL.String())
			return
		}
		http.NotFound(w, r)
	})
}
