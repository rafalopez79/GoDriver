package config

import (
	"testing"

	"gotest.tools/assert"
)

func TestParse(t *testing.T) {
	txt := `{
	  "serverport": 8000,
	   "webport": 8080,
	   "connections": [
		{"id": "test1", "user": "user1", "password":"password1"},
		{"id": "test2", "user": "user2", "password":"password2"}
		]}`
	c, err := Parse([]byte(txt))
	if err != nil {
		t.Errorf("Error parsing json: %s", err)
	} else {
		assert.Equal(t, len(c.Connections), 2)
		assert.Equal(t, c.Connections[0].ID, "test1")
		assert.Equal(t, c.Connections[0].User, "user1")
		assert.Equal(t, c.Connections[1].Password, "password2")
	}
}
