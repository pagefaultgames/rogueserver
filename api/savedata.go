package api

import "net/http"

// /api/savedata/get - get save data

type SavedataGetRequest struct{}
type SavedataGetResponse struct{}

func HandleSavedataGet(w http.ResponseWriter, r *http.Request) {

}

// /api/savedata/update - update save data

type SavedataUpdateRequest struct{}
type SavedataUpdateResponse struct{}

func HandleSavedataUpdate(w http.ResponseWriter, r *http.Request) {

}

// /api/savedata/delete - delete save date

type SavedataDeleteRequest struct{}
type SavedataDeleteResponse struct{}

func HandleSavedataDelete(w http.ResponseWriter, r *http.Request) {

}
