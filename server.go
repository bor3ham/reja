package reja

import (
	"fmt"
	"database/sql"
	"github.com/bor3ham/reja/models"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	db *sql.DB
	defaultDirectPageSize int
	maximumDirectPageSize int
	indirectPageSize int
	models map[string]*models.Model
}

func New(db *sql.DB) *Server {
	return &Server{
		db: db,
		defaultDirectPageSize: 50,
		maximumDirectPageSize: 100,
		indirectPageSize: 10,
		models: map[string]*models.Model{},
	}
}

func (s *Server) GetDatabase() *sql.DB {
	return s.db
}

func (s *Server) GetDefaultDirectPageSize() int {
	return s.defaultDirectPageSize
}
func (s *Server) SetDefaultDirectPageSize(size int) {
	s.defaultDirectPageSize = size
}

func (s *Server) GetMaximumDirectPageSize() int {
	return s.maximumDirectPageSize
}
func (s *Server) SetMaximumDirectPageSize(size int) {
	s.maximumDirectPageSize = size
}

func (s *Server) GetIndirectPageSize() int {
	return s.indirectPageSize
}
func (s *Server) SetIndirectPageSize(size int) {
	s.indirectPageSize = size
}

func (s *Server) RegisterModel(model *models.Model) {
	_, exists := s.models[model.Type]
	if exists {
		panic(fmt.Sprintf("Model %s already registered!", model.Type))
	}
	s.models[model.Type] = model
}
func (s *Server) GetModel(modelType string) *models.Model {
	mt, exists := s.models[modelType]
	if !exists {
		return nil
	}
	return mt
}

func (s *Server) Handle(router *mux.Router, modelType string, path string) {
	model, exists := s.models[modelType]
	if !exists {
		panic(fmt.Sprintf("Model %s not found!", modelType))
	}

	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		model.ListHandler(s, w, r)
	})
	router.HandleFunc(path + "/", func(w http.ResponseWriter, r *http.Request) {
		model.ListHandler(s, w, r)
	})
	router.HandleFunc(path + "/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		model.DetailHandler(s, w, r)
	})
	router.HandleFunc(path + "/{id:[0-9]+}/", func(w http.ResponseWriter, r *http.Request) {
		model.DetailHandler(s, w, r)
	})
}
