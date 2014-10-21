package server

import (
	"encoding/json"
	"net/http"
	"gopkg.in/mgo.v2/bson"
	"gitlab.mitre.org/fhir/models"
	"github.com/gorilla/mux"
	"os"
)

func RelatedPersonIndexHandler(rw http.ResponseWriter, r *http.Request) {
	var result []models.RelatedPerson
	c := Database.C("relatedpersons")
	iter := c.Find(nil).Limit(100).Iter()
	err := iter.All(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(result)
}

func RelatedPersonShowHandler(rw http.ResponseWriter, r *http.Request) {

	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	}	else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := Database.C("relatedpersons")

	result := models.RelatedPerson{}
	err := c.Find(bson.M{"_id": id.Hex()}).One(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(result)
}

func RelatedPersonCreateHandler(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	relatedperson := &models.RelatedPerson{}
	err := decoder.Decode(relatedperson)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := Database.C("relatedpersons")
	i := bson.NewObjectId()
	relatedperson.Id = i.Hex()
	err = c.Insert(relatedperson)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	host, err := os.Hostname()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Add("Location", "http://" + host + "/relatedperson/" + i.Hex())
}

func RelatedPersonUpdateHandler(rw http.ResponseWriter, r *http.Request) {

	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	}	else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	decoder := json.NewDecoder(r.Body)
	relatedperson := &models.RelatedPerson{}
	err := decoder.Decode(relatedperson)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := Database.C("relatedpersons")
	relatedperson.Id = id.Hex()
	err = c.Update(bson.M{"_id": id.Hex()}, relatedperson)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func RelatedPersonDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	}	else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := Database.C("relatedpersons")

	err := c.Remove(bson.M{"_id": id.Hex()})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

}