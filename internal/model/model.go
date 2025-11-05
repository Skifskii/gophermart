package model

import "time"

type UserAuth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Status string

const (
	StatusNew        Status = "NEW"        // заказ загружен в систему, но не попал в обработку
	StatusProcessing Status = "PROCESSING" // вознаграждение за заказ рассчитывается
	StatusInvalid    Status = "INVALID"    // система расчёта вознаграждений отказала в расчёте
	StatusProcessed  Status = "PROCESSED"  // данные по заказу проверены и информация о расчёте успешно
)

type Order struct {
	Number     string
	Status     Status
	Accrual    *float64
	UploadedAt time.Time
	UserLogin  string
}
