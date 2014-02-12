package app

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp(t *testing.T) {
	a := &Account{}
	a.Create(&NewAccountForm{
		"wsc",
		"balls",
		"balls"})
	log.Println(a)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", "/balls", nil)
	if err != nil {
		panic(err)
	}
	AccountForm(w, r)
	log.Println(w.Body.String())
}
