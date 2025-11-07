package withdrawal

import (
	"errors"
	"gophermart/internal/model"
	"gophermart/internal/service/orders"
	"time"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type WithdrawalManager struct {
	repo Repository
	bm   balanceManager
	om   orderManager
}

type Repository interface {
	RecordWithdrawal(withdrawal model.Withdrawal) error
	GetUserWithdrawals(userLogin string) ([]model.Withdrawal, error)
}

func New(repo Repository, bm balanceManager, om orderManager) *WithdrawalManager {
	return &WithdrawalManager{
		repo: repo,
		bm:   bm,
		om:   om,
	}
}

type balanceManager interface {
	GetBalance(login string) (current, withdrawn float64, err error)
	UpdateBalance(login string, amount float64) error
}

type orderManager interface {
	GetOrder(orderNum string) (model.Order, error)
}

func (wm *WithdrawalManager) RecordWithdrawal(userLogin, orderNum string, amount float64) error {
	// Проверяем, что у пользователя достаточно средств
	cur, _, err := wm.bm.GetBalance(userLogin)
	if err != nil {
		return err
	}
	if cur+amount < 0 {
		return ErrInsufficientFunds
	}

	// Проверяем заказ
	order, err := wm.om.GetOrder(orderNum)
	if err != nil {
		return err
	}
	if order.UserLogin != userLogin {
		return orders.ErrAnotherUserUploadedThisOrder
	}

	// Списываем средства
	if err := wm.bm.UpdateBalance(userLogin, amount); err != nil {
		return err
	}

	// Записываем списание
	return wm.repo.RecordWithdrawal(model.Withdrawal{
		Order:       orderNum,
		Sum:         amount * -1,
		ProcessedAt: time.Now(),
		UserLogin:   userLogin,
	})
}
