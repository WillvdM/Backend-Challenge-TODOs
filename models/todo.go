//Seperate request data from DB data

package models 
import "time"

type TodoInput struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Completed   bool   `json:"completed"`
    CreatedAt time.Time `json: "created at"`
}
