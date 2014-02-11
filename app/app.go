package app

import ()

type Account struct {
	Username string `dynamodb:"_hash"`
	Password string `dynamodb:"password"`
}
