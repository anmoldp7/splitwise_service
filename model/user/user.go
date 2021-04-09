package user

type User struct {
	UserID       string `json:"user_id"`
	Name         string `json:"name"`
	Password     string `json:"password"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	RegisteredOn string `json:"registered_on"`
	LastActiveOn string `json:"last_active_on"`
}
