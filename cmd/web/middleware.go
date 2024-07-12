package main

import "net/http"

func LoadSessions(next http.Handler) http.Handler {
	return sess.LoadAndSave(next)
}
