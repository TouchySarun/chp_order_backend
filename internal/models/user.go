package models

import "encoding/json"

type User struct {
	Id       string   `firestore:"id,omitempty" json:"id"`
	Username string   `firestore:"username" json:"username"`
	Password string   `json:"-"` // omit in JSON
	Name     string   `firestore:"name" json:"name"`
	Branch   string   `firestore:"branch" json:"branch"`
	Role     string   `firestore:"role" json:"role"`
	Ap       []string `firestore:"ap" json:"ap"`  // Ensure Ap is a slice
	Rack     string   `firestore:"rack" json:"rack"`
}
func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User
	if u.Ap == nil {
		u.Ap = []string{}
	}
	return json.Marshal((*Alias)(u))
}