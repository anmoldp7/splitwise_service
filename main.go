package main

import (
	"log"
	"net/http"

	"splitwise.com/m/controller"
	"splitwise.com/m/db"
)

func init() {
	db.Initialize()
}

func main() {
	defer db.CleanUp()
	http.HandleFunc("/addUser", controller.AddUser)
	http.HandleFunc("/addTransaction", controller.AddTransaction)
	http.HandleFunc("/getUserInfo", controller.GetUserInfo)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("web server crashed with error: %s", err.Error())
	}
}
