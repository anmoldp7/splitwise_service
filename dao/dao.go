package dao

import (
	"fmt"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"splitwise.com/m/db"
	"splitwise.com/m/model/userResponse"
)

func AddUser(name, password, phone, email string) (string, bool) {
	currentTime := time.Now()
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		log.Println("password encryption failed")
		return err.Error(), false
	}

	userID, isValid := validateCredentialsAndGetUniqueUserID(email, phone)
	if !isValid {
		log.Println("failed to add user, email or phone already registered")
		return "failed to add user, email or phone already registered", false
	}

	_, err = db.DB.Exec(
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

	err := db.DB.QueryRow(
		`SELECT name,phone,email,registered_on,last_active_on FROM service_user WHERE service_user_id=$1`,
		userID,
	).Scan(&name, &phone, &email, &registeredOn, &lastActiveOn)

	if err != nil {
		return nil, false
	}

	var debtAmount float64
	err = db.SumQuery(&debtAmount, `SELECT SUM(amount) FROM debt WHERE borrower=$1`, userID)
	if err != nil {
		log.Println("FAILED TO PERFORM INNER JOIN")
		log.Printf("Error: %s\n", err.Error())
		return nil, false
	}

	var loanAmount float64
	err = db.SumQuery(&loanAmount, `SELECT SUM(amount) FROM debt WHERE lender=$1`, userID)
	if err != nil {
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
		LoanedAmount: loanAmount,
		DebtAmount:   debtAmount,
	}, true
}

func validateUser(userID string) bool {
	return db.RowExists(`SELECT service_user_id FROM service_user WHERE service_user_id=$1`, userID)
}

func updateUserLastSeenOn(userID string, lastSeenOn time.Time) bool {
	_, err := db.DB.Exec(
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

func addDebtForIndividual(lender, borrower string, amount float64, submittedOn time.Time) bool {
	transactionID := getUniqueTransactionID()
	_, err := db.DB.Exec(
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

func validateCredentialsAndGetUniqueUserID(email, phone string) (string, bool) {
	isCredentialValid := !db.RowExists(`SELECT service_user_id FROM service_user WHERE email=$1 OR phone=$2`, email, phone)
	if !isCredentialValid {
		return "", false
	}

	var isValidID = false
	var userID string
	for !isValidID {
		userID = uuid.NewV4().String()
		isValidID = !db.RowExists(`SELECT service_user_id FROM service_user WHERE service_user_id=$1`, userID)
	}
	return userID, true
}

func getUniqueTransactionID() string {
	var isValidID = false
	var transactionID string
	for !isValidID {
		transactionID = uuid.NewV4().String()
		isValidID = !db.RowExists(`SELECT transaction_id FROM debt WHERE transaction_id=$1`, transactionID)
	}
	return transactionID
}
