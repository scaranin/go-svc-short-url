package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *URLHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", contentTypeTextPlain)
	var (
		cookieW *http.Cookie
		err     error
	)
	cookieR, err := r.Cookie(h.Auth.CookieName)
	if err != nil {
		log.Print(err.Error())
	}
	cookieW, err = h.Auth.FillUserReturnCookie(cookieR)
	if err != nil {
		log.Print(err.Error())
	}
	if err == http.ErrNoCookie {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if len(h.TrustedSubnet) == 0 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if !h.CheckIP(r) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	stat, err := h.Storage.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf, err := json.Marshal(stat)
	if err != nil {
		http.SetCookie(w, cookieW)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.SetCookie(w, cookieW)
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}
