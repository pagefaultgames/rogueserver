package api

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Flashfyre/pokerogue-server/db"
	"github.com/klauspost/compress/zstd"
)

const sessionSlotCount = 3

// /savedata/get - get save data

func (s *Server) HandleSavedataGet(w http.ResponseWriter, r *http.Request) {
	uuid, err := GetUuidFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch r.URL.Query().Get("datatype") {
	case "0": // System
		system, err := GetSystemSaveData(uuid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		saveJson, err := json.Marshal(system)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal save to json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write(saveJson)
	case "1": // Session
		slotId, err := strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert slot id: %s", err), http.StatusBadRequest)
			return
		}

		if slotId < 0 || slotId >= sessionSlotCount {
			http.Error(w, fmt.Sprintf("slot id %d out of range", slotId), http.StatusBadRequest)
			return
		}

		session, err := GetSessionSaveData(uuid, slotId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		saveJson, err := json.Marshal(session)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal save to json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write(saveJson)
	default:
		http.Error(w, "invalid data type", http.StatusBadRequest)
		return
	}
}

// /savedata/update - update save data

func (s *Server) HandleSavedataUpdate(w http.ResponseWriter, r *http.Request) {
	uuid, err := GetUuidFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	hexUuid := hex.EncodeToString(uuid)

	switch r.URL.Query().Get("datatype") {
	case "0": // System
		var system SystemSaveData
		err = json.NewDecoder(r.Body).Decode(&system)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request body: %s", err), http.StatusBadRequest)
			return
		}

		if system.TrainerId == 0 && system.SecretId == 0 {
			http.Error(w, "invalid system data", http.StatusInternalServerError)
			return
		}

		var gobBuffer bytes.Buffer
		err = gob.NewEncoder(&gobBuffer).Encode(system)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to serialize save: %s", err), http.StatusInternalServerError)
			return
		}

		zstdWriter, err := zstd.NewWriter(nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create zstd writer, %s", err), http.StatusInternalServerError)
			return
		}

		compressed := zstdWriter.EncodeAll(gobBuffer.Bytes(), nil)

		err = os.MkdirAll("userdata/"+hexUuid, 0755)
		if err != nil && !os.IsExist(err) {
			http.Error(w, fmt.Sprintf("failed to create userdata folder: %s", err), http.StatusInternalServerError)
			return
		}

		err = os.WriteFile("userdata/"+hexUuid+"/system.pzs", compressed, 0644)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write save file: %s", err), http.StatusInternalServerError)
			return
		}
	case "1": // Session
		slotId, err := strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert slot id: %s", err), http.StatusBadRequest)
			return
		}

		if slotId < 0 || slotId >= sessionSlotCount {
			http.Error(w, fmt.Sprintf("slot id %d out of range", slotId), http.StatusBadRequest)
			return
		}

		fileName := "session"
		if slotId != 0 {
			fileName += strconv.Itoa(slotId)
		}

		var session SessionSaveData
		err = json.NewDecoder(r.Body).Decode(&session)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request body: %s", err), http.StatusBadRequest)
			return
		}

		var gobBuffer bytes.Buffer
		err = gob.NewEncoder(&gobBuffer).Encode(session)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to serialize save: %s", err), http.StatusInternalServerError)
			return
		}

		zstdWriter, err := zstd.NewWriter(nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create zstd writer, %s", err), http.StatusInternalServerError)
			return
		}

		compressed := zstdWriter.EncodeAll(gobBuffer.Bytes(), nil)

		err = os.MkdirAll("userdata/"+hexUuid, 0755)
		if err != nil && !os.IsExist(err) {
			http.Error(w, fmt.Sprintf("failed to create userdata folder: %s", err), http.StatusInternalServerError)
			return
		}

		err = os.WriteFile(fmt.Sprintf("userdata/%s/%s.pzs", hexUuid, fileName), compressed, 0644)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write save file: %s", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "invalid data type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// /savedata/delete - delete save data

func (s *Server) HandleSavedataDelete(w http.ResponseWriter, r *http.Request) {
	uuid, err := GetUuidFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	hexUuid := hex.EncodeToString(uuid)

	switch r.URL.Query().Get("datatype") {
	case "0": // System
		err := os.Remove("userdata/" + hexUuid + "/system.pzs")
		if err != nil && !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("failed to delete save file: %s", err), http.StatusInternalServerError)
			return
		}
	case "1": // Session
		slotId, err := strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert slot id: %s", err), http.StatusBadRequest)
			return
		}

		if slotId < 0 || slotId >= sessionSlotCount {
			http.Error(w, fmt.Sprintf("slot id %d out of range", slotId), http.StatusBadRequest)
			return
		}

		fileName := "session"
		if slotId != 0 {
			fileName += strconv.Itoa(slotId)
		}

		err = os.Remove(fmt.Sprintf("userdata/%s/%s.pzs", hexUuid, fileName))
		if err != nil && !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("failed to delete save file: %s", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "invalid data type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type SavedataClearResponse struct {
	Success bool `json:"success"`
}

// /savedata/clear - mark session save data as cleared and delete

func (s *Server) HandleSavedataClear(w http.ResponseWriter, r *http.Request) {
	uuid, err := GetUuidFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	slotId, err := strconv.Atoi(r.URL.Query().Get("slot"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to convert slot id: %s", err), http.StatusBadRequest)
		return
	}

	if slotId < 0 || slotId >= sessionSlotCount {
		http.Error(w, fmt.Sprintf("slot id %d out of range", slotId), http.StatusBadRequest)
		return
	}

	session, err := GetSessionSaveData(uuid, slotId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionCompleted := ValidateSessionCompleted(session)
	newCompletion := false

	if sessionCompleted {
		newCompletion, err = db.TryAddSeedCompletion(uuid, session.Seed, int(session.GameMode))
		if err != nil {
			log.Print("failed to mark seed as completed")
		}
	}

	response, err := json.Marshal(SavedataClearResponse{Success: newCompletion})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	if sessionCompleted {
		fileName := "session"
		if slotId != 0 {
			fileName += strconv.Itoa(slotId)
		}

		err = os.Remove(fmt.Sprintf("userdata/%s/%s.pzs", hex.EncodeToString(uuid), fileName))
		if err != nil && !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("failed to delete save file: %s", err), http.StatusInternalServerError)
			return
		}
	}

	w.Write(response)
}
