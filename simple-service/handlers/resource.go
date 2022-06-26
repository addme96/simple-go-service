//go:generate mockgen -destination=mocks/resource.go -package mocks . ResourceRepository
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/addme96/simple-go-service/simple-service/entities"
	"github.com/go-chi/chi/v5"
)

type ResourceRepository interface {
	Create(ctx context.Context, newResource entities.Resource) (int, error)
	Read(ctx context.Context, id int) (*entities.Resource, error)
	ReadAll(ctx context.Context) ([]entities.Resource, error)
	Update(ctx context.Context, id int, newResource entities.Resource) error
	Delete(ctx context.Context, id int) error
}

type Resource struct {
	Repository ResourceRepository
}

func NewResource(repository ResourceRepository) *Resource {
	return &Resource{Repository: repository}
}

func (r *Resource) Post(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Content-Type") != "application/json" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	var newResource entities.Resource
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(bytes, &newResource); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	var id int
	if id, err = r.Repository.Create(request.Context(), newResource); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusCreated)
	resp := fmt.Sprintf(`{"id": %d}`, id)
	writer.Write([]byte(resp))
}

func (r *Resource) GetCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		resourceID := chi.URLParam(request, "resourceID")
		ID, err := strconv.Atoi(resourceID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		resource, err := r.Repository.Read(request.Context(), ID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(request.Context(), "resource", resource)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

var getFromCtxError = errors.New("failed to read resource from the context")

func (r *Resource) Get(writer http.ResponseWriter, request *http.Request) {
	resource, ok := request.Context().Value("resource").(*entities.Resource)
	if !ok {
		http.Error(writer, getFromCtxError.Error(), http.StatusBadRequest)
		return
	}
	bytes, _ := json.Marshal(resource)
	writer.Write(bytes)
}

func (r *Resource) List(writer http.ResponseWriter, request *http.Request) {
	resources, err := r.Repository.ReadAll(request.Context())
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, _ := json.Marshal(resources)
	writer.Write(bytes)
}

func (r *Resource) Put(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Content-Type") != "application/json" {
		http.Error(writer, "invalid Content-Type - should be application/json", http.StatusBadRequest)
		return
	}
	currentResource, ok := request.Context().Value("resource").(*entities.Resource)
	if !ok {
		http.Error(writer, getFromCtxError.Error(), http.StatusBadRequest)
		return
	}
	var newResource entities.Resource
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(bytes, &newResource); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if err = r.Repository.Update(request.Context(), currentResource.ID, newResource); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *Resource) Delete(writer http.ResponseWriter, request *http.Request) {
	currentResource, ok := request.Context().Value("resource").(*entities.Resource)
	if !ok {
		http.Error(writer, getFromCtxError.Error(), http.StatusBadRequest)
		return
	}
	err := r.Repository.Delete(request.Context(), currentResource.ID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}
