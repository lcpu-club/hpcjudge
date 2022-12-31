package common

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

type CommonServer struct {
	listen    string
	mux       *http.ServeMux
	signed    bool
	secretKey []byte
}

func NewCommonServer(listen string, secretKey []byte) *CommonServer {
	signed := secretKey != nil
	return &CommonServer{
		listen:    listen,
		mux:       http.NewServeMux(),
		signed:    signed,
		secretKey: secretKey,
	}
}

func (cs *CommonServer) GetMux() *http.ServeMux {
	return cs.mux
}

func (cs *CommonServer) RespondError(w http.ResponseWriter, statusCode int, message string) {
	if message == "" {
		message = strconv.Itoa(statusCode) + " " + http.StatusText(statusCode)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func (cs *CommonServer) ParseRequest(w http.ResponseWriter, r *http.Request, req Request) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		cs.RespondError(w, 400, "")
		return false
	}
	if cs.signed {
		signature := []byte(r.Header.Get("X-Signature"))
		signed, err := CheckSignedMessage(body, cs.secretKey, signature)
		if err != nil {
			log.Println("ERROR:", err)
			cs.RespondError(w, 500, "")
			return false
		}
		if !signed {
			cs.RespondError(w, 403, "")
			return false
		}
	}
	err = json.Unmarshal(body, req)
	if err != nil {
		cs.RespondError(w, 400, "")
		return false
	}
	return true
}

func (cs *CommonServer) Respond(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(resp)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}
	_, err = w.Write(j)
	if err != nil {
		log.Println("ERROR:", err)
	}
}

func (cs *CommonServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cs.mux.ServeHTTP(w, r)
}

func (cs *CommonServer) Start() error {
	return http.ListenAndServe(cs.listen, cs)
}
