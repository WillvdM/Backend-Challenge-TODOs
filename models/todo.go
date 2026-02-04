//Seperate request data from DB data

package models

import "time"

type TodoInput struct {
	Title     string    `json:"title,omitempty"` // Leave field if it is empty
	Completed bool      `json:"completed,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
