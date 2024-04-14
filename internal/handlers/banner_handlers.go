package handlers

import (
	"banner-service/internal/middlewares"
	"banner-service/internal/models"
	bannerservice "banner-service/internal/services"
	"banner-service/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

type BannerHandler struct {
	bannerService *bannerservice.BannerService
}

func NewBannerHandler(service *bannerservice.BannerService) *BannerHandler {
	return &BannerHandler{
		bannerService: service,
	}
}

func InitBannerRoutes(bannerService *bannerservice.BannerService, r *mux.Router) {
	bh := NewBannerHandler(bannerService)

	s := r.PathPrefix("/auth").Subrouter()

	s.Use(middlewares.AuthMiddleware)
	s.HandleFunc("/banners", bh.GetBannersHandler).Methods("GET")
	s.HandleFunc("/banner", bh.GetBannerHandler).Methods("GET")
	s.HandleFunc("/banner", bh.CreateBannerHandler).Methods("POST")
	s.HandleFunc("/banner/{id}", bh.UpdateBannerHandler).Methods("PUT")
	s.HandleFunc("/banner/{id}", bh.DeleteBannerHandler).Methods("DELETE")

}

func (h *BannerHandler) GetBannerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	tagIDStr := r.URL.Query().Get("tag_id")
	featureIDStr := r.URL.Query().Get("feature_id")
	useLastRevisionStr := r.URL.Query().Get("use_last_revision")
	isAdmin, ok := r.Context().Value("isAdminKey").(bool)

	if !ok {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil || tagID <= 0 {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	featureID, err := strconv.Atoi(featureIDStr)
	if err != nil || featureID <= 0 {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	useLastRevision, err := strconv.ParseBool(useLastRevisionStr)
	if err != nil {
		useLastRevision = false
	}

	banner, err := h.bannerService.GetBanner(ctx, tagID, featureID, useLastRevision, isAdmin)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			http.Error(w, "Баннер для не найден", http.StatusNotFound)
		} else {
			println(err.Error())
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	err = json.NewEncoder(w).Encode(banner)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

func (h *BannerHandler) GetBannersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	isAdmin, ok := r.Context().Value("isAdminKey").(bool)

	if !ok {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if !isAdmin {
		http.Error(w, "Пользователь не имеет доступа", http.StatusForbidden)
		return
	}

	tagID, err := utils.ParsePositiveInt(r.URL.Query().Get("tag_id"))
	if err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	featureID, err := utils.ParsePositiveInt(r.URL.Query().Get("feature_id"))
	if err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	limit, err := utils.ParsePositiveInt(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	offset, err := utils.ParsePositiveInt(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	banners, err := h.bannerService.GetBanners(ctx, featureID, tagID, limit, offset)
	if err != nil {
		println(err.Error())
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(banners)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

func (h *BannerHandler) CreateBannerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	isAdmin, ok := r.Context().Value("isAdminKey").(bool)

	if !ok {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Error(w, "Пользователь не имеет доступа", http.StatusForbidden)
		return
	}
	var banner models.Banner
	if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
		println(err.Error())
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	id, err := h.bannerService.CreateBanner(ctx, &banner)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(struct {
		BannerID int `json:"banner_id"`
	}{
		BannerID: id,
	})
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

func (h *BannerHandler) UpdateBannerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	isAdmin, ok := r.Context().Value("isAdminKey").(bool)

	if !ok {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Error(w, "Пользователь не имеет доступа", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bannerIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	var banner models.Banner
	if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	if err := h.bannerService.UpdateBanner(ctx, bannerID, &banner); err != nil {
		if err.Error() == "no rows affected" {
			http.Error(w, "Баннер не найден", http.StatusNotFound)
			return
		}
		if err.Error() == "ERROR: Not a unique combination of tag_id and feature_id (SQLSTATE P0001)" {
			http.Error(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		println(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

func (h *BannerHandler) DeleteBannerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	isAdmin, ok := r.Context().Value("isAdminKey").(bool)

	if !ok {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Error(w, "Пользователь не имеет доступа", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bannerIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	if err := h.bannerService.DeleteBanner(ctx, bannerID); err != nil {
		if err.Error() == "no rows affected" {
			http.Error(w, "Баннер не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
