package bendis

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (b *Bendis) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	if b.Debug {
		mux.Use(middleware.Logger)
	}
	mux.Use(middleware.Recoverer)
	mux.Use(b.SessionLoad)
	mux.Use(b.NoSurf)
	mux.Use(b.CheckForMaintenanceMode)

	//mux.Get("/", func(w http.ResponseWriter, r *http.Request){
	//	fmt.Fprint(w, "Welcome to Bendis")
	//})

	return mux
}

// Routes are Bendis specific routes, which are mounted in the routes file
// in Bendis applications
func Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/test-c", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("it works!"))
	})
	return r
}
