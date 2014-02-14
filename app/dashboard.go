package app

import (
	"net/http"
)

var Dashboard = &Resource{
	Path: "/u/",
	Get:  Authed(GetDashboard)}

func GetDashboard(r *http.Request, e *Env, m map[string]interface{}) (Response, error) {
	data := struct{ Username string }{
		m["account"].(*Account).Username}
	return Anonymous("dashboard", data), nil
}
