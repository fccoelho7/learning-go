package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

type Todo struct {
	gorm.Model
	Id			uint		`gorm:"primaryKey"`
	Description	string		`gorm:"not null"`
	Completed   bool
	CreatedAt   time.Time	`gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time	`gorm:"default:CURRENT_TIMESTAMP"`
}

var db, dbErr = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

func healthChecker(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	todo := &Todo{}
	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println("Unable to read the body: %v\n", err)
	 }

	json.Unmarshal(reqBody, &todo)

	db.Create(&todo)

	json.NewEncoder(w).Encode(todo)
}

func allItems(w http.ResponseWriter, r *http.Request) {
	todos := []Todo{}

	result := db.Find(&todos).Error

	if result != nil {
	   log.Println("Unable to delete todo: %v\n", result)
	}
	
	json.NewEncoder(w).Encode(todos)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	todo := Todo{}

	db.Limit(1).Find(&todo, params["id"])

	json.NewEncoder(w).Encode(todo)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	todo := Todo{}

	db.First(&todo, params["id"])

	db.Delete(&todo)

	json.NewEncoder(w).Encode(todo)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	todo := Todo{}

	db.First(&todo, params["id"])

	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println("Unable to read the body: %v\n", err)
	 }

	json.Unmarshal(reqBody, &todo)

	db.Save(&todo)

	json.NewEncoder(w).Encode(todo)
}

func main() {
	if dbErr != nil {
		panic("Failed to connect database")
	}

	db.AutoMigrate(&Todo{})

	log.Info("Starting Todolist API server")

	router := mux.NewRouter()

	router.HandleFunc("/health_checker", healthChecker).Methods("GET")
	router.HandleFunc("/todo", createItem).Methods("POST")
	router.HandleFunc("/todos", allItems).Methods("GET")
	router.HandleFunc("/todo/{id}", getItem).Methods("GET")
	router.HandleFunc("/todo/{id}", updateItem).Methods("PUT")
	router.HandleFunc("/todo/{id}", deleteItem).Methods("DELETE")
	
	http.ListenAndServe(":8000", router)
}