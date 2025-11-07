package balance

type BalanceManager struct {
	repo Repository
}

type Repository interface {
	GetBalance(login string) (current, withdrawn float64, err error)
	UpdateBalance(login string, amount float64) error
}

func New(repo Repository) *BalanceManager {
	return &BalanceManager{repo: repo}
}

func (bm *BalanceManager) GetBalance(login string) (current, withdrawn float64, err error) {
	return bm.repo.GetBalance(login)
}

func (bm *BalanceManager) UpdateBalance(login string, amount float64) error {
	return bm.repo.UpdateBalance(login, amount)
}
