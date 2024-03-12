/*
Copyright © 2024 Fabien
*/

package users

// GlobalUsersResponse
// GET /api/get-global-users
type GlobalUsersResponse struct {
	Data []GlobalUsersData `json:"data"`
}

type GlobalUsersData struct {
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	IsAdmin       bool   `json:"isAdmin"`
}

type AccountResponse struct {
	ID   string `json:"sub"`
	Name string `json:"name"`
}
