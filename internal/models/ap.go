package models

type Ap struct {
	Code string `json:"code" firestore:"code"`
	Name string `json:"name" firestore:"name"`
	Rack string `json:"rack" firestore:"rack"`
}