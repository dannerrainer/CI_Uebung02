package ci_uebung02

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8010", a.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id}
	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	p := product{ID: id}
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getRating(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["rating_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid rating ID")
		return
	}

	r := rating{RatingId: id}
	if err := r.getRating(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Rating not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, r)
}

func (a *App) getRatingsForProduct(w http.ResponseWriter, req *http.Request) {
	count, _ := strconv.Atoi(req.FormValue("count"))
	start, _ := strconv.Atoi(req.FormValue("start"))

	vars := mux.Vars(req)
	productId, err := strconv.Atoi(vars["product_id"])

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	ratings, err := getRatingsForProduct(a.DB, productId, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, ratings)
}

func (a *App) createRating(w http.ResponseWriter, req *http.Request) {
	var r rating
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&r); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()

	if err := r.createRating(a.DB); err != nil {
		// hideous and not recommended way to check for error!!!
		if err.Error() == "pq: insert or update on table \"ratings\" violates foreign key constraint \"ratings_product_id_fkey\"" {
			respondWithError(w, http.StatusBadRequest, "No product with the specified id exists!")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, r)
}

func (a *App) updateRating(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["rating_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid rating ID")
		return
	}

	var r rating
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&r); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	r.RatingId = id

	if err := r.updateRating(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, r)
}

func (a *App) deleteRating(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["rating_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid rating ID")
		return
	}

	r := rating{RatingId: id}
	if err := r.deleteRating(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")

	a.Router.HandleFunc("/rating", a.createRating).Methods("POST")
	a.Router.HandleFunc("/rating/{rating_id:[0-9]+}", a.getRating).Methods("GET")
	a.Router.HandleFunc("/ratings/{product_id:[0-9]+}", a.getRatingsForProduct).Methods("GET")
	a.Router.HandleFunc("/rating/{rating_id:[0-9]+}", a.updateRating).Methods("PUT")
	a.Router.HandleFunc("/rating/{rating_id:[0-9]+}", a.deleteRating).Methods("DELETE")
}
