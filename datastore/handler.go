/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package datastore

import (
	"encoding/json"
	"fmt"
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/husobee/vestigo"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

var datastoreRepo Repository = datastoreMgoRepo{}

func GetItems(w http.ResponseWriter, r *http.Request) {

	dataItems, err := datastoreRepo.FindAll()
	if err != nil {
		log.Err(err).Msg("cannot retrieve data items")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputil.WriteResponse(w, r, toDataItemsResponse(r, dataItems))
}

func GetItem(w http.ResponseWriter, r *http.Request) {

	key := vestigo.Param(r, "key")
	dataItem, err := datastoreRepo.Get(key)
	if err != nil {
		switch err {
		case dataItemNotFound:
			log.Err(err).Msgf("Data item key=%s not found", key)
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Err(err).Msgf("Cannot retrieve data item key=%s", key)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(httputil.HeaderContentType, dataItem.ContentType)
	w.Write(dataItem.Value)
}

func StoreItem(w http.ResponseWriter, r *http.Request) {
	item, err := toDataItem(r)
	if err != nil {
		log.Err(err).Msg("Error storing data store item")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := datastoreRepo.Store(item)
	if err != nil {
		log.Err(err).Msgf("Cannot store item key=%s", item.Key)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if updated {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {

	key := vestigo.Param(r, "key")
	if err := datastoreRepo.Remove(key); err != nil {
		switch err {
		case dataItemNotFound:
			log.Err(err).Msgf("Data item key=%s not found", key)
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Err(err).Msgf("Cannot delete item key=%s", key)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	log.Info().Msgf("Deleted data item key=%s", key)
	w.WriteHeader(http.StatusNoContent)
}

func GetDataStoreValue(key string) (interface{}, error) {

	dataItem, err := datastoreRepo.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot find datastore item key=%s: %v", key, err)
	}

	if !strings.HasPrefix(dataItem.ContentType, httputil.MediaTypeJson) &&
		!strings.HasPrefix(dataItem.ContentType, "text/json") {
		return string(dataItem.Value), nil
	}

	value := map[string]interface{}{}
	if err := json.Unmarshal(dataItem.Value, &value); err != nil {
		return nil, fmt.Errorf("cannot unmarshal datastore item key=%s: %v", key, err)
	}
	return value, nil
}
