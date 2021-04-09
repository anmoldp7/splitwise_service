package userResponse

type UserResponse struct {
	UserID       string  `json:"user_id"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`
	Email        string  `json:"email"`
	RegisteredOn string  `json:"registered_on"`
	LastActiveOn string  `json:"last_active_on"`
	LoanedAmount float64 `json:"loaned_amount"`
	DebtAmount   float64 `json:"debt_amount"`
}
