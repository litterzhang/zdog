package test

import (
	"testing"
	"fmt"
	"zdog/model"
)

func TestGetUser(t *testing.T) {
	user := model.GetUserByName("zhang")
	fmt.Println(user)
}

func TestCreateUser(t *testing.T) {
	user := model.CreateUser("zhang", "199408")
	fmt.Println(user)
}
