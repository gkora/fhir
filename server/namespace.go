package server

import (
	"encoding/json"
	"net/http"
	"gopkg.in/mgo.v2/bson"
	"gitlab.mitre.org/fhir/models"
	"github.com/gorilla/mux"
	"os"
)

func NamespaceIndexHandler(rw http.ResponseWriter, r *http.Request) {
	var result []models.Namespace
	c := Database.C("namespaces")
	iter := c.Find(nil).Limit(100).Iter()
	err := iter.All(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(result)
}

func NamespaceShowHandler(rw http.ResponseWriter, r *http.Request) {

	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	}	else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := Database.C("namespaces")

	result := models.Namespace{}
	err := c.Find(bson.M{"_id": id.Hex()}).One(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(result)
}

func NamespaceCreateHandler(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	namespace := &models.Namespace{}
	err := decoder.Decode(namespace)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := Database.C("namespaces")
	i := bson.NewObjectId()
	namespace.Id = i.Hex()
	err = c.Insert(namespace)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	host, err := os.Hostname()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Add("Location", "http://" + host + "/namespace/" + i.Hex())
}

func NamespaceUpdateHandler(rw http.ResponseWriter, r *http.Request) {

	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	}	else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	decoder := json.NewDecoder(r.Body)
	namespace := &models.Namespace{}
	err := decoder.Decode(namespace)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := Database.C("namespaces")
	namespace.Id = id.Hex()
	err = c.Update(bson.M{"_id": id.Hex()}, namespace)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func NamespaceDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	}	else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := Database.C("namespaces")

	err := c.Remove(bson.M{"_id": id.Hex()})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

}