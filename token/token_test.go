package token

import (
	"testing"
	"time"
)

const (
	t_password = "my super strong password"
)

func TestTokenizationWorks(t *testing.T) {

	j := NewTerminateJob("us-west-2", "1234")

	token, err := Tokenize(t_password, j)
	if err != nil {
		t.Error(err)
	}

	j2, err2 := Untokenize(t_password, token)

	if err2 != nil {
		t.Error(err2)
	}

	j1j := string(j.JSON())
	j2j := string(j2.JSON())

	if j1j != j2j {
		t.Error("Tokenization integrity failed")
	}
}

func TestTokenizationFailsHMAC(t *testing.T) {
	j := NewTerminateJob("us-west-2", "1234")

	token, _ := Tokenize(t_password, j)

	_t := []byte(token)
	_t[2] = _t[2] + 1 // just change a bit somewhere

	token = string(_t)

	_, err := Untokenize(t_password, token)

	if err == nil {
		t.Error("error expected")
	}
}

func TestJobTokenExpired(t *testing.T) {

	j := &JobToken{
		ValidUntil: time.Now().Add(time.Duration(-1 * time.Hour)),
	}

	if !j.Expired() {
		t.Error("expected expired token")
	}
}
