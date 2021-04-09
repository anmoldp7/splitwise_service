package transaction

type Transaction struct {
	Debtor string   `json:"debtor"`
	Group  []string `json:"group"`
	Amount float64  `json:"amount"`
}
