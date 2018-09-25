package model

import (
	"reflect"
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	Name     string `json:"name" form:"name"`
	Password string `json:"password,omitempty" form:"password"`
	Token    string `json:"token,omitempty" form:"token,omitempty"`
}

var (
	inited = false
)

const (
	user_bucket = "user"
	jwt_key = "xxx"
)

func bindIndex() {
	if inited {
		return
	}
	BindIndex(user_bucket, "name")
	inited = true
}


func check(kvs map[string]interface{}) bool {
	if v, in := kvs["name"]; !in || reflect.TypeOf(v).Kind() != reflect.String || len(v.(string)) == 0{
		return false
	}
	if v, in := kvs["password"]; !in || reflect.TypeOf(v).Kind() != reflect.String || len(v.(string)) == 0{
		return false
	}
	return true
}

func (u *User) ToJson() map[string]interface{} {
	userJson := map[string]interface{} {
		"name" : u.Name,
		"password" : u.Password,
	}
	return userJson
}

func (u *User) Check() bool {
	return check(u.ToJson())
}

func (u *User) Find() *User {
	if u == nil {
		return nil
	}
	user := GetUserByName(u.Name)
	return user
}

func (u *User) Save() {
	if u == nil {
		return
	}
	CreateUser(u.Name, u.Password)
}

func (u *User) GenToken() (_token string) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = u.Name
	claims["password"] = u.Password

	_token, err := token.SignedString([]byte(jwt_key))	
	if err != nil {
		return
	}
	u.Token = _token
	return _token
}

func UserFromToken(_token string) *User {
	if len(_token) == 0 {
		return nil
	}
	token, err := jwt.Parse(_token, func(token *jwt.Token) (interface{}, error) {
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(jwt_key), nil
	})
	if err != nil {
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		name := claims["name"]
		password := claims["password"]

		if name == nil || password == nil {
			return nil
		}
		return &User {
			Name: name.(string),
			Password: password.(string),
			Token: _token,
		}
	}
	return nil
}

func GetUserByName(name string) *User {
	if len(name) == 0 {
		return nil
	}
	bindIndex()
	
	// query 
	querys := map[string]interface{} {
		"name": name,
	}
	userJson := Get(user_bucket, querys)
	if userJson == nil || !check(userJson) {
		return nil
	}
	return &User {
		Name: userJson["name"].(string),
		Password: userJson["password"].(string),
	}
}

func CreateUser(name string, password string) *User {
	user := &User {
		Name: name,
		Password: password,
	}
	userJson := user.ToJson()
	Insert(user_bucket, userJson)
	return user
}

