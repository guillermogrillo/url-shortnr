package api

import (
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"strings"
	"urlshortn/pkg/hash"
	"urlshortn/pkg/storage"
	"urlshortn/pkg/token"
)

type HttpUrlHandler interface {
	ShortenUrl(http.ResponseWriter, *http.Request)
	GetLongUrl(http.ResponseWriter, *http.Request)
	DeleteShortenUrl(http.ResponseWriter, *http.Request)
}

type UrlHandler struct {
	TokenGen    token.TokenGenerator
	TokenHasher hash.TokenHasher
	UrlStore    storage.Store
	logger      *slog.Logger
}

func NewUrlHandler(tokenGen token.TokenGenerator, urlTokenHasher hash.TokenHasher, urlStore storage.Store, logger *slog.Logger) UrlHandler {
	return UrlHandler{
		TokenGen:    tokenGen,
		TokenHasher: urlTokenHasher,
		UrlStore:    urlStore,
		logger:      logger,
	}
}

type ShortenUrlRequest struct {
	URL string `json:"url"`
}

type ShortenUrlResponse struct {
	ShortUrl string `json:"short_url"`
}

func (h *UrlHandler) ShortenUrl(w http.ResponseWriter, r *http.Request) {
	var req ShortenUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error decoding the request to a known struct", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(struct {
			Error string
		}{"invalid request"})
		return
	}
	h.logger.Debug("Shortening url", "url", req.URL)

	token, err := h.TokenGen.GenerateToken()
	if err != nil {
		h.logger.Error("Error generating a token based on the url", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct {
			Error string
		}{"internal error generating a token"})
		return
	}
	h.logger.Debug("Generated token", "token", token)

	shortenUrl, err := h.TokenHasher.Hash(int64(token))
	if err != nil {
		h.logger.Error("Error generating a hash for the token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct {
			Error string
		}{"internal error generating a hash for the token"})
		return
	}
	h.logger.Debug("Generated shorten url", "url", shortenUrl)

	err = h.UrlStore.Store(shortenUrl, req.URL)
	if err != nil {
		h.logger.Error("Error persisting in storage", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct {
			Error string
		}{"internal error persisting a token"})
		return
	}

	result := ShortenUrlResponse{
		ShortUrl: shortenUrl,
	}

	response, err := json.Marshal(result)
	if err != nil {
		h.logger.Error("Error marshalling the response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct {
			Error string
		}{"internal error generating the response"})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

func (h *UrlHandler) GetLongUrl(w http.ResponseWriter, r *http.Request) {
	shortenUrl := strings.TrimPrefix(r.URL.Path, "/shortn/")
	if shortenUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{Error: "no shortenUrl provided"})
		return
	}
	h.logger.Debug("GetLongURl", "url", shortenUrl)
	longUrl, err := h.UrlStore.Fetch(shortenUrl)
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			h.logger.Error("Provided short url not found in redis", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Error string
			}{"the provided short url is not available"})
			return
		default:
			h.logger.Error("Error fetching long url from storage", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct {
				Error string
			}{"internal error getting the long url"})
			return
		}
	}

	http.Redirect(w, r, longUrl, http.StatusFound)
}

func (h *UrlHandler) DeleteShortenUrl(w http.ResponseWriter, r *http.Request) {
	shortenUrl := strings.TrimPrefix(r.URL.Path, "/shortn/")
	if shortenUrl == "" {
		h.logger.Error("No shortenUrl provided")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{Error: "no shortenUrl provided"})
		return
	}
	h.logger.Debug("DeleteShortenUrl", "url", shortenUrl)
	err := h.UrlStore.Remove(shortenUrl)
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			h.logger.Error("Provided short url not found in redis", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Error string
			}{"the provided short url is not available"})
			return
		default:
			h.logger.Error("Error deleting short url from storage", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct {
				Error string
			}{"internal error deleting the short url"})
			return
		}
	}
}
