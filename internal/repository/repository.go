package repository

import "errors"

var ErrLoginAlreadyTaken = errors.New("a user with this login already exists")
var ErrOrderAlreadyExists = errors.New("an order with this number already exists")
var ErrWithdrawalAlreadyExists = errors.New("a withdrawal with this number already exists")
var ErrUserLoginNotFound = errors.New("a user with this login not found")
var ErrOrderNotFound = errors.New("an order with this number not found")
var ErrWrongPassword = errors.New("wrong password")
