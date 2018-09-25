package test

import (
	"testing"
	"zdog/model"
)

func TestBoltBindIndex(t *testing.T) {
	model.BindIndex("user", "name")
}

func TestBoltInsert(t *testing.T) {
	model.Insert("user", map[string]interface{} {
		"name": "zhang",
		"password": "19940824",
	})
}

func TestBoltQuery(t *testing.T) {
	model.Query("user", map[string]interface{} {
		"name": "zhang",
		"password": "19940824",
	})
}

func TestMain(m *testing.M) {
	model.OpenDb()
    m.Run()
	model.CloseDb()
}
