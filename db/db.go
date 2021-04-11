package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
	"splitwise.com/m/config"
)

var DB *sql.DB
var initialiseOnce sync.Once
var cleanupOnce sync.Once

func Initialize() {
	initialiseOnce.Do(func() {
		log.Println("INIT DB CALLED:")
		initializeDB()
	})
}

func CleanUp() {
	cleanupOnce.Do(func() {
		log.Println("DB connection closed")
		DB.Close()
	})
}
func initializeDB() {
	connectToDB()
	pingDB()
	createUserTable()
	createDebtTable()
}

func connectToDB() {
	psqlConn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.HOST,
		config.PORT,
		config.USER,
		config.PASSWORD,
		config.DB_NAME,
	)

	var err error
	DB, err = sql.Open("postgres", psqlConn)
	if err != nil {
		log.Println("FAILED TO CONNECT TO DB")
		log.Fatalf("Error: %s\n", err.Error())
	}
	log.Printf("CONNECTED TO %s DB\n", config.DB_NAME)
}

func pingDB() {
	err := DB.Ping()
	if err != nil {
		log.Println("FAILED TO PING DB")
		log.Fatalf("Error: %s\n", err.Error())
	}
	log.Printf("DB PING SUCCESS")
}

func createUserTable() {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS service_user(
			service_user_id VARCHAR(36) PRIMARY KEY,
			name TEXT NOT NULL,
			password TEXT NOT NULL,
			phone TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			registered_on TIMESTAMP,
			last_active_on TIMESTAMP
		)`)

	if err != nil {
		log.Println("FAILED TO CREATE TABLE SERVICE USER")
		log.Fatalf("Error: %s\n", err.Error())
	}
	log.Printf("USERS TABLE CREATED")

}

func createDebtTable() {
	_, err := DB.Exec(`
	CREATE TABLE IF NOT EXISTS debt(
		transaction_id VARCHAR(36) PRIMARY KEY,
		lender VARCHAR(36) NOT NULL REFERENCES service_user(service_user_id) ON DELETE CASCADE ON UPDATE CASCADE,
		borrower VARCHAR(36) NOT NULL REFERENCES service_user(service_user_id) ON DELETE CASCADE ON UPDATE CASCADE,
		amount NUMERIC,
		submitted_on TIMESTAMP
	)`)

	if err != nil {
		log.Println("FAILED TO CREATE TABLE DEBT")
		log.Fatalf("Error: %s\n", err.Error())
	}
	log.Printf("DEBT TABLE CREATED")

}

func RowExists(query string, args ...interface{}) bool {
	var rowExists bool
	query = fmt.Sprintf("SELECT EXISTS(%s)", query)
	err := DB.QueryRow(query, args...).Scan(&rowExists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("DUPLICATE DB SCAN FOR QUERY: %s FAILED\n", query)
	}
	return rowExists
}

func SumQuery(amount *float64, query string, args ...interface{}) error {
	var debtAmount sql.NullFloat64
	err := DB.QueryRow(query, args...).Scan(&debtAmount)
	if err != sql.ErrNoRows && err != nil {
		log.Println("FAILED TO PERFORM INNER JOIN")
		log.Printf("Error: %s\n", err.Error())
		return err
	}
	*amount = debtAmount.Float64
	return nil
}
