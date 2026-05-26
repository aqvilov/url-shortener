package alias

//GENERATE ALIAS (save-хендлер)

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"url-shortener/lib/random"
)

type Request struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type Response struct {
	OriginalUrl string `json:"original_url"`
	Alias       string `json:"alias"`
	RespStatus  status
}

type status struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type URLSaver interface {
	SaveUrl(alias string, originalUrl string) error
}

const LenAliasIfEmpty = 6

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(slog.String("op", "handler.alias.New")) // к каждому логу будем добавлять маркер (то, откуда он)

		var req Request

		err := json.NewDecoder(r.Body).Decode(&req) // декодируем из тела запроса в req
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(Response{
					RespStatus: status{Status: "Error", Error: "empty request body"},
				})
				return
			}

			log.Error("failed to decode request body", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				RespStatus: status{Status: "Error", Error: "failed to decode json"},
			})
			return
		}

		if req.URL == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				RespStatus: status{Status: "Error", Error: "url is required"},
			})
			return
		}

		if _, err := url.ParseRequestURI(req.URL); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				RespStatus: status{Status: "Error", Error: "invalid url format"},
			})
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(LenAliasIfEmpty) // генерация если пусто
		}

		err = urlSaver.SaveUrl(alias, req.URL)
		if err != nil {
			log.Error("failed to save url to database", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				RespStatus: status{Status: "Error", Error: "internal database error"},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		StatusOK(w, r, alias)
	}
}

func StatusOK(w http.ResponseWriter, r *http.Request, alias string) {
	json.NewEncoder(w).Encode(Response{
		Alias: alias,
		RespStatus: status{
			Status: "OK",
		},
	})
}
