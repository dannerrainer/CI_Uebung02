package ci_uebung02_test

import (
	"bytes"
	"encoding/json"
	ci_uebung02 "github.com/dannerrainer/CI_Uebung02.git"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a ci_uebung02.App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTableExists()
	code := m.Run()
	clearTables()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableProductionsCreation); err != nil {
		log.Fatal(err)
	}
	if _, err := a.DB.Exec(tableRatingsCreation); err != nil {
		log.Fatal(err)
	}
}

func clearTables() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")

	// postgres needs to be owner of table AND sequence!
	a.DB.Exec("DELETE FROM ratings")
	a.DB.Exec("ALTER SEQUENCE ratings_rating_id_seq RESTART WITH 1")
}

const tableProductionsCreation = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`
const tableRatingsCreation = `CREATE TABLE IF NOT EXISTS ratings
(
        rating_id SERIAL PRIMARY KEY,
        product_id INTEGER REFERENCES products(id),
        rating INTEGER NOT NULL,
        info TEXT NULL
)`

func TestEmptyTable(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {

	clearTables()

	var jsonStr = []byte(`{"name":"test product", "price": 11.22}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTables()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func TestUpdateProduct(t *testing.T) {

	clearTables()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	var jsonStr = []byte(`{"name":"test product - updated name", "price": 11.22}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTables()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestEmptyRatingsTable(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/ratings/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentRating(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/rating/123", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Rating not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Rating not found'. Got '%s'", m["error"])
	}
}

func TestCreateRating(t *testing.T) {
	clearTables()

	var jsonStr = []byte(`{"name":"test product", "price": 11.22}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}

	// test new rating here
	jsonStr = []byte(`{"product_id":1, "rating": 7, "rating_text": "Das isch jutesch Zeuch!"}`)
	req, _ = http.NewRequest("POST", "/rating", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m2 map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m2)

	if m2["rating_text"] != "Das isch jutesch Zeuch!" {
		t.Errorf("Expected rating text to be 'Das isch jutesch Zeuch!'. Got '%v'", m2["rating_text"])
	}

	if m2["rating"] != 7.0 {
		t.Errorf("Expected rating to be 7. Got %v", m2["rating"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m2["rating_id"] != 1.0 {
		t.Errorf("Expected product ID to be 1. Got %v", m2["rating_id"])
	}
}
func TestCreateRatingWithoutProduct(t *testing.T) {
	clearTables()

	jsonStr := []byte(`{"product_id":1, "rating": 7, "rating_text": "Das isch jutesch Zeuch!"}`)
	req, _ := http.NewRequest("POST", "/rating", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)

	var m2 map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m2)

	if m2["error"] != "No product with the specified id exists!" {
		t.Errorf("Expected error message to be 'No product with the specified id exists!'. Got '%v'", m2["error"])
	}
}

func TestGetRating(t *testing.T) {
	clearTables()
	addRatings(1)

	req, _ := http.NewRequest("GET", "/rating/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addRatings(count int) {
	if count < 1 {
		count = 1
	}

	a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(1), (1+1.0)*10)
	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO ratings(product_id, rating, info) VALUES($1, $2, $3)", 1, i, "Static rating text...")
	}
}

func TestUpdateRating(t *testing.T) {
	clearTables()
	addProducts(2)
	addRatings(5)

	req, _ := http.NewRequest("GET", "/rating/5", nil)
	response := executeRequest(req)
	var originalRating map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRating)

	var jsonStr = []byte(`{"product_id":2, "rating": 0, "rating_text": "Some updated rating text"}`)
	req, _ = http.NewRequest("PUT", "/rating/5", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["rating_id"] != originalRating["rating_id"] {
		t.Errorf("Expected the rating_id to remain the same (%v). Got %v", originalRating["rating_id"], m["rating_id"])
	}

	if m["product_id"] == originalRating["product_id"] {
		t.Errorf("Expected changed product_id from (%v) to (%v). Got %v", originalRating["product_id"], m["product_id"], m["product_id"])
	}

	if m["rating"] == originalRating["rating"] {
		t.Errorf("Expected the rating to change from '%v' to '%v'. Got '%v'", originalRating["rating"], m["rating"], m["rating"])
	}

	if m["rating_text"] == originalRating["rating_text"] {
		t.Errorf("Expected the rating_text to change from '%v' to '%v'. Got '%v'", originalRating["rating_text"], m["rating_text"], m["rating_text"])
	}
}

func TestDeleteRating(t *testing.T) {
	clearTables()
	addRatings(1)

	req, _ := http.NewRequest("GET", "/rating/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/rating/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/rating/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
