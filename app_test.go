package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	err := a.Initialise(DbUser, DbPassword, "test")
	if err != nil {
		log.Fatal("Error occurred while initialising the database")
	}
	createTable()
	m.Run()
}

func createTable() {
	createTableQuery := `create table if not exists products(
		id int NOT NULL AUTO_INCREMENT,
		name varchar(255) NOT NULL,
		quantity int,
		price float(10,7),
		PRIMARY KEY(id)
		);`

	_, err := a.DB.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("Delete from products")
	a.DB.Exec("Alter table products AUTO_INCREMENT=1")
	log.Println("Clear table")
}

func addProduct(name string, quantity int, price float64) {
	query := fmt.Sprintf("insert into products(name,quantity,price) values('%v',%v,%v)", name, quantity, price)
	a.DB.Exec(query)
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProduct("keyboard", 100, 500)
	request, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(request)
	checkStatusCode(t, http.StatusOK, response.Code)
}

func checkStatusCode(t *testing.T, expectedStatusCode int, actualStatusCode int) {
	if expectedStatusCode != actualStatusCode {
		t.Errorf("Expected staus : %v Received : %v", expectedStatusCode, actualStatusCode)
	}
}

func sendRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, request)
	return recorder
}

func TestCreateProduct(t *testing.T) {
	clearTable()
	var product = []byte(`{"name":"chair", "quantity":5, "price":100}`)
	request, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(product))
	request.Header.Set("Content-Type", "application/json")
	response := sendRequest(request)
	checkStatusCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "chair" {
		t.Errorf("Expected name : %v but Got : %v", "chair", m["name"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProduct("connector", 10, 100)
	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProduct("connector", 10, 100)
	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	var oldVal map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &oldVal)

	var product = []byte(`{"name":"connector", "quantity":5, "price":100}`)
	request, _ := http.NewRequest("PUT", "/product/1", bytes.NewBuffer(product))
	request.Header.Set("Content-Type", "application/json")
	response = sendRequest(request)

	var newVal map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &newVal)

	if oldVal["id"] != newVal["id"] {
		t.Errorf("Expected id: %v Got id: %v", newVal["id"], oldVal["id"])
	}
}
