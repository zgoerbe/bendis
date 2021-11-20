package bendis

import "net/http"

func (b *Bendis) SessionLoad(next http.Handler) http.Handler{
	b.InfoLog.Println("SessionLoad called")
	return b.Session.LoadAndSave(next)
}