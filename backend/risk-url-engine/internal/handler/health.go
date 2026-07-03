package handler

import "net/http"

func Health(w http.ResponseWriter, r *http.Request){
	if r.Method!=http.MethodGet{
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type","text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}