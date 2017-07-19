package servers

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	db schema.Database

	defaultDirectPageSize int
	maximumDirectPageSize int
	indirectPageSize      int

	models map[string]schema.Model
	routes map[string]string

	whitespace bool
	easyJSON   bool
}

func New(db schema.Database) *Server {
	return &Server{
		db: db,

		defaultDirectPageSize: 50,
		maximumDirectPageSize: 100,
		indirectPageSize:      10,

		models: map[string]schema.Model{},
		routes: map[string]string{},

		whitespace: false,
		easyJSON:   false,
	}
}

func (s *Server) EnableEasyJSON() {
	s.easyJSON = true
}
func (s *Server) DisableEasyJSON() {
	s.easyJSON = false
}
func (s *Server) UseEasyJSON() bool {
	return s.easyJSON
}

func (s *Server) EnableWhitespace() {
	s.whitespace = true
}
func (s *Server) DisableWhitespace() {
	s.whitespace = false
}
func (s *Server) Whitespace() bool {
	return s.whitespace
}

func (s *Server) GetDatabase() schema.Database {
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

func (s *Server) RegisterModel(model *schema.Model) {
	_, exists := s.models[model.Type]
	if exists {
		panic(fmt.Sprintf("Model %s already registered!", model.Type))
	}
	s.models[model.Type] = *model
}
func (s *Server) GetModel(modelType string) *schema.Model {
	mt, exists := s.models[modelType]
	if !exists {
		return nil
	}
	return &mt
}
func (s *Server) GetRoute(modelType string) string {
	path, exists := s.routes[modelType]
	if !exists {
		panic(fmt.Sprintf("Route for model %s not found!", modelType))
	}
	return path
}

func (s *Server) Handle(router *mux.Router, modelType string, path string) {
	model, exists := s.models[modelType]
	if !exists {
		panic(fmt.Sprintf("Model %s not found!", modelType))
	}
	route, exists := s.routes[modelType]
	if exists {
		panic(fmt.Sprintf("Model %s already registered at path %s!", modelType, route))
	}
	s.routes[modelType] = path

	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ListHandler(s, &model, w, r)
	})
	router.HandleFunc(path+"/", func(w http.ResponseWriter, r *http.Request) {
		ListHandler(s, &model, w, r)
	})
	router.HandleFunc(path+"/filters", func(w http.ResponseWriter, r *http.Request) {
		FilterInfoHandler(s, &model, w, r)
	})
	router.HandleFunc(path+"/filters/", func(w http.ResponseWriter, r *http.Request) {
		FilterInfoHandler(s, &model, w, r)
	})
	router.HandleFunc(path+"/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		DetailHandler(s, &model, w, r)
	})
	router.HandleFunc(path+"/{id:[0-9]+}/", func(w http.ResponseWriter, r *http.Request) {
		DetailHandler(s, &model, w, r)
	})

	for _, relationship := range model.Relationships {
		relation := relationship
		router.HandleFunc(
			path+"/{id:[0-9]+}/relationships/"+relation.GetKey(),
			func(w http.ResponseWriter, r *http.Request) {
				RelationHandler(s, &model, relation, w, r)
			},
		)
		router.HandleFunc(
			path+"/{id:[0-9]+}/relationships/"+relation.GetKey()+"/",
			func(w http.ResponseWriter, r *http.Request) {
				RelationHandler(s, &model, relation, w, r)
			},
		)
	}
}
