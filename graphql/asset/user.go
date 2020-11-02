package asset

import (
	"encoding/json"
	"fmt"
	"github.com/mainflux/mainflux/graphql"
)

var (
	users  []User
	phones []string
)

type userData struct {
	Users   []User   `json:"User"`
}

type User struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Phone   string   `json:"phone"`
}

func initUser() {
	result := graphql.Query("queryUsers")

	var userData userData
	e := json.Unmarshal(result, &userData)
	fmt.Println("userData:",  userData)
	if e != nil {

	}

	//userData.Users[0].Phone = "18758183504"
	users = userData.Users
	fmt.Println("users len:", len(users))
}


func GetPhones() []string {
	if len(phones) == 0 {
		for i := 0; i < len(users) ; i++  {
			// 未做手机号码去重
			if len(users[i].Phone) == 11 {
				phones = append(phones, users[i].Phone)
			}
		}
	}
	return phones
}