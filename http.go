package reja

import (
	"github.com/bor3ham/reja/models"
	"github.com/gorilla/mux"
)

func RegisterHandlers(router *mux.Router, model models.Model, path string) {
	router.HandleFunc(path, model.ListHandler)
	router.HandleFunc(path+"/", model.ListHandler)
	router.HandleFunc(path+"/{id:[0-9]+}", model.DetailHandler)
	router.HandleFunc(path+"/{id:[0-9]+}/", model.DetailHandler)
}
