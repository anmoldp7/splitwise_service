package transaction

type Transaction struct {
	Lender string   `json:"lender"`
	Borrowers  []string `json:"borrowers"`
	Amount float64  `json:"amount"`
}
