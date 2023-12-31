package api

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/klauspost/compress/zstd"
)

// /savedata/get - get save data

func (s *Server) HandleSavedataGet(w http.ResponseWriter, r *http.Request) {
	uuid, err := GetUuidFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hexUuid := hex.EncodeToString(uuid)

	switch r.URL.Query().Get("datatype") {
	case "0": // System
		save, err := os.ReadFile("userdata/" + hexUuid + "/system.pzs")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read save file: %s", err), http.StatusInternalServerError)
			return
		}

		zstdReader, err := zstd.NewReader(nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create zstd reader: %s", err), http.StatusInternalServerError)
			return
		}

		decompressed, err := zstdReader.DecodeAll(save, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decompress save file: %s", err), http.StatusInternalServerError)
			return
		}

		gobDecoderBuf := bytes.NewBuffer(decompressed)

		var system SystemSaveData
		err = gob.NewDecoder(gobDecoderBuf).Decode(&system)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to deserialize save: %s", err), http.StatusInternalServerError)
			return
		}

		saveJson, err := json.Marshal(system)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal save to json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write(saveJson)
	case "1": // Session
		save, err := os.ReadFile("userdata/" + hexUuid + "/session.pzs")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read save file: %s", err), http.StatusInternalServerError)
			return
		}

		zstdReader, err := zstd.NewReader(nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create zstd reader: %s", err), http.StatusInternalServerError)
			return
		}

		decompressed, err := zstdReader.DecodeAll(save, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decompress save file: %s", err), http.StatusInternalServerError)
			return
		}

		gobDecoderBuf := bytes.NewBuffer(decompressed)

		var session SessionSaveData
		err = gob.NewDecoder(gobDecoderBuf).Decode(&session)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to deserialize save: %s", err), http.StatusInternalServerError)
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

	hexUuid := hex.EncodeToString(uuid)

	switch r.URL.Query().Get("datatype") {
	case "0": // System
		var system SystemSaveData
		err = json.NewDecoder(r.Body).Decode(&system)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request body: %s", err), http.StatusBadRequest)
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
		if !os.IsExist(err) {
			http.Error(w, fmt.Sprintf("failed to create userdata folder: %s", err), http.StatusInternalServerError)
			return
		}

		err = os.WriteFile("userdata/"+hexUuid+"/system.pzs", compressed, 0644)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write save file: %s", err), http.StatusInternalServerError)
			return
		}
	case "1": // Session
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
		if !os.IsExist(err) {
			http.Error(w, fmt.Sprintf("failed to create userdata folder: %s", err), http.StatusInternalServerError)
			return
		}

		err = os.WriteFile("userdata/"+hexUuid+"/session.pzs", compressed, 0644)
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

	hexUuid := hex.EncodeToString(uuid)

	switch r.URL.Query().Get("datatype") {
	case "0": // System
		err := os.Remove("userdata/"+hexUuid+"/system.pzs")
		if !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("failed to delete save file: %s", err), http.StatusInternalServerError)
			return 
		}
	case "1": // Session
		err := os.Remove("userdata/"+hexUuid+"/session.pzs")
		if !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("failed to delete save file: %s", err), http.StatusInternalServerError)
			return 
		}
	default:
		http.Error(w, "invalid data type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
