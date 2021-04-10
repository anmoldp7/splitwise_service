package controller

import (
	"encoding/json"
	"net/http"

	"splitwise.com/m/db"
	"splitwise.com/m/model/transaction"
	"splitwise.com/m/model/user"
	"splitwise.com/m/utils"
)

func AddUser(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res := map[string]string{"error": "method not allowed"}
		encodedRes, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, string(encodedRes), http.StatusMethodNotAllowed)
		return
	}

	var userData user.User
	json.NewDecoder(req.Body).Decode(&userData)

	if !utils.ValidateRequest(
		userData.Name,
		userData.Password,
		userData.Phone,
		userData.Email,
	) {
		res := map[string]string{"error": "invalid input"}
		encodedRes, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, string(encodedRes), http.StatusUnprocessableEntity)
		return
	}

	userID, isValid := db.AddUser(userData.Name, userData.Password, userData.Phone, userData.Email)
	if !isValid {
		errMsg := userID
		res := map[string]string{"error": errMsg}
		encodedRes, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, string(encodedRes), http.StatusUnprocessableEntity)
		return
	}

	res := map[string]string{
		"user_id": userID,
		"name":    userData.Name,
		"email":   userData.Email,
		"phone":   userData.Phone,
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func AddTransaction(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res := map[string]string{"error": "method not allowed"}
		encodedRes, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, string(encodedRes), http.StatusMethodNotAllowed)
		return
	}

	var transactionData transaction.Transaction
	json.NewDecoder(req.Body).Decode(&transactionData)

	transactionID, isValid := db.AddDebt(transactionData.Lender, transactionData.Borrowers, transactionData.Amount)
	if !isValid {
		errMsg := transactionID
		res := map[string]string{"error": errMsg}
		encodedRes, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, string(encodedRes), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactionData)
}

func GetUserInfo(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res := map[string]string{"error": "method not allowed"}
		encodedRes, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, string(encodedRes), http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseForm()
	if err != nil {
		res := map[string]string{"error": "unprocessable entity"}
		encodedRes, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, string(encodedRes), http.StatusUnprocessableEntity)
		return
	}

	userID := req.FormValue("userID")

	userInfo, isValid := db.GetUserDebtAndLoaned(userID)
	if !isValid {
		res := map[string]string{"error": "unprocessable entity"}
		encodedRes, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, string(encodedRes), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)

}
