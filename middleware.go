package bendis

import (
	"fmt"
	"github.com/justinas/nosurf"
	"net/http"
	"strconv"
	"strings"
)

func (b *Bendis) SessionLoad(next http.Handler) http.Handler {
	b.InfoLog.Println("SessionLoad called")
	return b.Session.LoadAndSave(next)
}

func (b *Bendis) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(b.config.cookie.secure)

	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		Path:     "/",
		Domain:   b.config.cookie.domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return csrfHandler
}

func (b *Bendis) CheckForMaintenanceMode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if maintenanceMode {
			if !strings.Contains(request.URL.Path, "/public/maintenance.html") {
				writer.WriteHeader(http.StatusServiceUnavailable)
				writer.Header().Set("Retry-After:", "300")
				writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
				http.ServeFile(writer, request, fmt.Sprintf("%s/public/maintenance.html", b.RootPath))
				return
			}
		}
		next.ServeHTTP(writer, request)
	})
}
