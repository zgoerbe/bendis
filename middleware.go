package bendis

import (
	"github.com/justinas/nosurf"
	"net/http"
	"strconv"
)

func (b *Bendis) SessionLoad(next http.Handler) http.Handler{
	b.InfoLog.Println("SessionLoad called")
	return b.Session.LoadAndSave(next)
}

func (b *Bendis) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(b.config.cookie.secure)

	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		Path:       "/",
		Domain:     b.config.cookie.domain,
		Secure:     secure,
		HttpOnly:   true,
		SameSite:   http.SameSiteStrictMode,
	})

	return csrfHandler
}