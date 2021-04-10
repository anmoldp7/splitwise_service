package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"splitwise.com/m/config"
	"splitwise.com/m/model/userResponse"
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

func AddUser(name, password, phone, email string) (string, bool) {
	currentTime := time.Now()
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		log.Println("password encryption failed")
		return err.Error(), false
	}
	userID := getUniqueUserID()
	_, err = DB.Exec(
		`INSERT INTO service_user(service_user_id, name, password, phone, email, registered_on, last_active_on) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID,
		name,
		string(encryptedPassword),
		phone,
		email,
		currentTime,
		currentTime,
	)

	if err != nil {
		log.Println("failed to add service_user in database")
		return err.Error(), false
	}

	return userID, true
}

func AddDebt(lender string, borrowers []string, amount float64) (string, bool) {
	if !validateUser(lender) {
		return "invalid lender service user ID", false
	}

	for _, borrower := range borrowers {
		if !validateUser(borrower) {
			return fmt.Sprintf("invalid borrower ID: %s", borrower), false
		}
	}
	debtAmount := amount / float64(len(borrowers))
	currentTime := time.Now()
	var isValid bool
	isValid = updateUserLastSeenOn(lender, currentTime)
	if !isValid {
		return "failed to update last seen status of lender", false
	}
	for _, borrower := range borrowers {
		isValid = addDebtForIndividual(lender, borrower, debtAmount, currentTime)
		if !isValid {
			log.Printf("FAILED TO ADD TRANSACTION FOR LENDER: %s\n", lender)
			return "failed to add transaction", false
		}
		isValid = updateUserLastSeenOn(borrower, currentTime)
		if !isValid {
			return "failed to update last seen status of borrower", false
		}
	}
	return "", true
}

func GetUserDebtAndLoaned(userID string) (*userResponse.UserResponse, bool) {

	var name, phone, email, registeredOn, lastActiveOn string

	err := DB.QueryRow(
		`SELECT name,phone,email,registered_on,last_active_on FROM service_user WHERE service_user_id=$1`,
		userID,
	).Scan(&name, &phone, &email, &registeredOn, &lastActiveOn)

	if err != nil {
		return nil, false
	}

	var debtAmount sql.NullFloat64
	err = DB.QueryRow(
		`SELECT SUM(amount) FROM debt WHERE borrower=$1`, userID).Scan(&debtAmount)
	if err != sql.ErrNoRows && err != nil {
		log.Println("FAILED TO PERFORM INNER JOIN")
		log.Printf("Error: %s\n", err.Error())
		return nil, false
	}

	var loanAmount sql.NullFloat64
	err = DB.QueryRow(
		`SELECT SUM(amount) FROM debt WHERE lender=$1`, userID).Scan(&debtAmount)
	if err != sql.ErrNoRows && err != nil {
		log.Println("FAILED TO PERFORM INNER JOIN")
		log.Printf("Error: %s\n", err.Error())
		return nil, false
	}
	return &userResponse.UserResponse{
		UserID:       userID,
		Name:         name,
		Phone:        phone,
		Email:        email,
		RegisteredOn: registeredOn,
		LastActiveOn: lastActiveOn,
		LoanedAmount: loanAmount.Float64,
		DebtAmount:   debtAmount.Float64,
	}, true
}

func addDebtForIndividual(lender, borrower string, amount float64, submittedOn time.Time) bool {
	transactionID := getUniqueTransactionID()
	_, err := DB.Exec(
		`INSERT INTO debt(transaction_id, lender, borrower, amount, submitted_on) VALUES($1, $2, $3, $4, $5)`,
		transactionID,
		lender,
		borrower,
		amount,
		submittedOn,
	)
	if err != nil {
		log.Printf("FAILED TO INSERT DEBT OF AMOUNT: %f, LENDER: %s, BORROWER: %s, ERROR: %s", amount, lender, borrower, err.Error())
		return false
	}
	return true
}

func updateUserLastSeenOn(userID string, lastSeenOn time.Time) bool {
	_, err := DB.Exec(
		`UPDATE service_user SET last_active_on=$1 WHERE service_user_id=$2`,
		lastSeenOn,
		userID,
	)
	if err != nil {
		log.Printf("FAILED TO UPDATE LAST ACTIVITY OF USER: %s\n", userID)
		return false
	}
	return true
}

func validateUser(userID string) bool {
	return rowExists(`SELECT service_user_id FROM service_user WHERE service_user_id=$1`, userID)
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
			phone TEXT NOT NULL,
			email TEXT NOT NULL,
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

func getUniqueUserID() string {
	var isValidID = false
	var userID string
	for !isValidID {
		userID = uuid.NewV4().String()
		isValidID = !rowExists(`SELECT service_user_id FROM service_user WHERE service_user_id=$1`, userID)
	}
	return userID
}

func getUniqueTransactionID() string {
	var isValidID = false
	var transactionID string
	for !isValidID {
		transactionID = uuid.NewV4().String()
		isValidID = !rowExists(`SELECT transaction_id FROM debt WHERE transaction_id=$1`, transactionID)
	}
	return transactionID
}

func rowExists(query string, args ...interface{}) bool {
	var rowExists bool
	query = fmt.Sprintf("SELECT EXISTS(%s)", query)
	err := DB.QueryRow(query, args...).Scan(&rowExists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("DUPLICATE DB SCAN FOR QUERY: %s FAILED\n", query)
	}
	return rowExists
}
