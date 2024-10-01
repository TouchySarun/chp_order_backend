package models

type User struct {
	Id	*string `json:"id"`
	Username string `firestore:"username" json:"username"`
	Password string  `firestore:"password" json:"password"`
	Name string `firestore:"name" json:"name"`
	Branch string `firestore:"branch" json:"branch"`
	Role string `firestore:"role" json:"role"`
	Ap []string `firestore:"ap" json:"ap"`
	Rack string `firestore:"rack" json:"rack"`
}