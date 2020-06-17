package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	
)


// The store variable is a package level variable that will be available for
// use throughout our application code
var store Store


func main() {

	fmt.Println("Starting server...")

	//connStr format should be for example: "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
	//see https://godoc.org/github.com/lib/pq for more details
	//connStr := "postgres://postgres_user:postgres@192.168.254.148/postgres_db"
	connString := os.Getenv("PG_CONNSTRING")
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfuly connected to databse!")
	InitStore(&dbStore{db: db})

	r := newRouter()
	fmt.Println("Serving on port 8080")
	http.ListenAndServe(":8080", r)
}

func helloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/hello", helloServer).Methods("GET")
	//	r.HandleFunc("/presents", handlers.GetPresentHandler).Methods("GET")
	//	r.HandleFunc("/presents", handlers.CreatePresentHandler).Methods("POST")

	r.HandleFunc("/presents", GetPresentHandler).Methods("GET")
	r.HandleFunc("/presents", CreatePresentHandler).Methods("POST")
	r.HandleFunc("/totalbudget", GetTotalBudgetHandler).Methods("GET")

	staticFileDirectory := http.Dir("./static/")

	staticFileHandler := http.StripPrefix("/static/", http.FileServer(staticFileDirectory))

	r.PathPrefix("/static/").Handler(staticFileHandler).Methods("GET")

	return r
}


//Store functions to interact wtih DB
func (store *dbStore) CreatePresent(present *Present) error {
	_, err := store.db.Query("INSERT INTO whishlist(person, present, budget) VALUES ($1, $2, $3)", present.Person, present.Present, present.Budget)
	return err
}

func (store *dbStore) GetPresent() ([]*Present, error) {
	rows, err := store.db.Query("SELECT person, present, budget FROM whishlist")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//Create data stuctre taht is returned by the function
	//It's empty by default

	presents := []*Present{}

	for rows.Next() {
		//creates a pointer for each returned row
		rowWithPresent := &Present{}
		if err := rows.Scan(&rowWithPresent.Person, &rowWithPresent.Present, &rowWithPresent.Budget); err != nil {
			return nil, err
		}
		//append the result to the returned array
		presents = append(presents, rowWithPresent)
	}
	return presents, nil
}

func (store *dbStore) GetTotalBudget() ([]*TotalBudget, error) {
	rows, err := store.db.Query("SELECT SUM (budget) AS Total from whishlist")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalBudget := []*TotalBudget{}

	for rows.Next() {
		rowWithBudget := &TotalBudget{}
		if err := rows.Scan(&rowWithBudget.Total); err != nil {
			return nil, err
		}
		totalBudget = append(totalBudget, rowWithBudget)
	}
	return totalBudget, nil
}

//PART OF STORE.GO 
/*
We will need to call the InitStore method to initialize the store. This will
typically be done at the beginning of our application (in this case, when the server starts up)
This can also be used to set up the store as a mock, which we will be observing
later on
*/

//var store Store

func InitStore(s Store) {
	store = s
}

//Store defines thre methdods, to Create and to Get presents and calculate TotalBudget
type Store interface {
	CreatePresent(present *Present) error
	GetPresent() ([]*Present, error)
	GetTotalBudget() ([]*TotalBudget, error)
}

//dbStore struct implements Store interface. It takes sql.DB connections object to represent the database connection.
type dbStore struct {
	db *sql.DB
}


//PART OF HANDLERS.GO

//handlers functions

//GetPresentHandler used by main.go
func GetPresentHandler(w http.ResponseWriter, r *http.Request) {
	presents, err := store.GetPresent()

	presentsListBytes, err := json.Marshal(presents)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(presentsListBytes)
}

//GetTotalBudgetHandler used by main.go
func GetTotalBudgetHandler(w http.ResponseWriter, r *http.Request) {
	totalBudget, err := store.GetTotalBudget()

	totalBudgetListBytes, err := json.Marshal(totalBudget)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(totalBudgetListBytes)
}

//CreatePresentHandler used by main.go
func CreatePresentHandler(w http.ResponseWriter, r *http.Request) {
	present := Present{}

	err := r.ParseForm()

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	present.Budget = r.Form.Get("budget")
	present.Person = r.Form.Get("person")
	present.Present = r.Form.Get("present")

	err = store.CreatePresent(&present)
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/static/", http.StatusFound)
}

var presents []Present

/*
//GetPresent used by main.go
func GetPresent(w http.ResponseWriter, r *http.Request) {
	presentListBytes, err := json.Marshal(presents)
	//return error by printing to console in case of error
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//if ok, write the JSON list od dreams to the reponse
	w.Write(presentListBytes)
}

//CreatePresent used by main.go
func CreatePresent(w http.ResponseWriter, r *http.Request) {
	present := Present{}
	//parse the form values
	err := r.ParseForm()
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//if ok
	present.Person = r.Form.Get("person")
	present.Budget = r.Form.Get("budget")
	present.Present = r.Form.Get("present")

	presents = append(presents, present)

	http.Redirect(w, r, "/static/", http.StatusFound)
}

*/

//Present used in ...
type Present struct {
	Budget  string `json:"budget"`
	Person  string `json:"person"`
	Present string `json:"present"`
}

//TotalBudget used in ...
type TotalBudget struct {
	Total int64 `json:"total"`
}
