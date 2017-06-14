package reja

import (
	"github.com/gorilla/mux"
	"github.com/bor3ham/reja/models"
)

func RegisterHandlers(router *mux.Router, model models.Model, path string) {
	router.HandleFunc(path, model.ListHandler)
	router.HandleFunc(path+"/", model.ListHandler)
	router.HandleFunc(path+"/{id:[0-9]+}", model.DetailHandler)
	router.HandleFunc(path+"/{id:[0-9]+}/", model.DetailHandler)
}
