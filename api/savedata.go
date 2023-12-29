package api

import "net/http"

// /savedata/get - get save data

type SavedataGetRequest struct{}
type SavedataGetResponse struct{}

func (s *Server) HandleSavedataGet(w http.ResponseWriter, r *http.Request) {

}

// /savedata/update - update save data

type SavedataUpdateRequest struct{}
type SavedataUpdateResponse struct{}

func (s *Server) HandleSavedataUpdate(w http.ResponseWriter, r *http.Request) {

}

// /savedata/delete - delete save date

type SavedataDeleteRequest struct{}
type SavedataDeleteResponse struct{}

func (s *Server) HandleSavedataDelete(w http.ResponseWriter, r *http.Request) {

}
