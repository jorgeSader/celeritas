package devify

import "net/http"

func (d *Devify) SessionLoad(next http.Handler) http.Handler {
	return d.Session.LoadAndSave(next)
}
