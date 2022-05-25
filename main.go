package ci_uebung02

import "os"

func main() {
	a := App{}
	/*a.Initialize(
	os.Getenv("APP_DB_USERNAME"),
	os.Getenv("APP_DB_PASSWORD"),
	os.Getenv("APP_DB_NAME"))
	*/
	a.Initialize(
		os.Getenv("postgres"),
		os.Getenv("postgres"),
		os.Getenv("ci_uebung02"))

	a.Run(":8010")
}
