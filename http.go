package reja

import (
  "github.com/gorilla/mux"
)

func RegisterHandlers(router *mux.Router, model Model, path string) {
  router.HandleFunc(path, model.ListHandler)
  router.HandleFunc(path+"/", model.ListHandler)
  router.HandleFunc(path+"/{id:[0-9]+}", model.DetailHandler)
  router.HandleFunc(path+"/{id:[0-9]+}/", model.DetailHandler)
}
