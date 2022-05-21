package ci_uebung02

import (
	"database/sql"
)

type rating struct {
	RatingId  int    `json:"rating_id"`
	ProductId int    `json:"product_id"`
	Rating    int    `json:"rating"`
	Text      string `json:"rating_text"`
}

func (rating *rating) createRating(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO ratings(product_id, rating, info) VALUES ($1, $2, $3) RETURNING rating_id",
		rating.ProductId, rating.Rating, rating.Text).Scan(&rating.RatingId)

	if err != nil {
		return err
	}

	return nil
}

func (rating *rating) getRating(db *sql.DB) error {
	return db.QueryRow("SELECT product_id, rating, info FROM ratings WHERE rating_id=$1",
		rating.RatingId).Scan(&rating.ProductId, &rating.Rating, &rating.Text)
}

func getRatingsForProduct(db *sql.DB, productId, start, count int) ([]rating, error) {
	rows, err := db.Query(
		"SELECT rating_id, product_id, rating, info FROM ratings WHERE product_id=$1 LIMIT $2 OFFSET $3",
		productId, count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// use this syntax to create empty array!
	ratings := []rating{}

	for rows.Next() {
		var r rating
		if err := rows.Scan(&r.RatingId, &r.ProductId, &r.Rating, &r.Text); err != nil {
			return nil, err
		}
		ratings = append(ratings, r)
	}

	return ratings, nil
}

func (r *rating) updateRating(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE ratings SET product_id=$1, rating=$2, info=$3 WHERE rating_id=$4",
			r.ProductId, r.Rating, r.Text, r.RatingId)

	return err
}

func (r *rating) deleteRating(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM ratings WHERE rating_id=$1", r.RatingId)

	return err
}
