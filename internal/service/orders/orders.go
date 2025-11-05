package orders

import (
	"errors"
	"gophermart/internal/model"
	"gophermart/internal/repository"
	"gophermart/internal/service/orders/loyalsys"
	"sort"
	"time"
)

var ErrWrongOrderNumFormat = errors.New("wrong order number format")
var ErrUserUploadedThisOrder = errors.New("the order number has already been uploaded by this user")
var ErrAnotherUserUploadedThisOrder = errors.New("the order number has already been uploaded by another user")
var errOrdersHaveDifferentNumbers = errors.New("orders with different numbers were received")

type OrdersManager struct {
	repo                 Repository
	infoGetter           OrderInfoGetter
	processingOrdersChan chan model.Order
	updatesChan          chan model.Order
}

type Repository interface {
	GetOrders(userLogin string) ([]model.Order, error)
	GetUnfinishedOrders() ([]model.Order, error)
	AddOrder(login string, order model.Order) error
	GetOrder(orderNum string) (model.Order, error)
	UpdateOrders(orders []model.Order) error
}

type OrderInfoGetter interface {
	GetOrderInfo(orderNum string) (model.Order, error)
}

func New(repo Repository, infoGetter OrderInfoGetter) *OrdersManager {
	om := &OrdersManager{
		repo:                 repo,
		infoGetter:           infoGetter,
		processingOrdersChan: make(chan model.Order, 100),
		updatesChan:          make(chan model.Order, 100),
	}

	go om.fillProcessingChanFromRepo()
	go om.checkOrderWorker()
	go om.updateRepoWorker()

	return om
}

func (om *OrdersManager) GetOrders(userLogin string) ([]model.Order, error) {
	orders, err := om.repo.GetOrders(userLogin)
	if err != nil {
		return nil, err
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.After(orders[j].UploadedAt)
	})

	return orders, nil
}

func (om *OrdersManager) AddOrder(userLogin, orderNum string) error {
	// Проверяем формат номера заказа
	if ok := validateOrderNumFormat(orderNum); !ok {
		return ErrWrongOrderNumFormat
	}

	// Ищем заказ среди сохраненных
	if ord, err := om.repo.GetOrder(orderNum); err == nil {
		// Если ошибок нет, значит заказ уже загружен в систему
		if ord.UserLogin == userLogin {
			return ErrUserUploadedThisOrder
		}
		return ErrAnotherUserUploadedThisOrder
	} else {
		// Ошибку отсутствия заказа пропускаем, остальные возвращаем
		if !errors.Is(err, repository.ErrOrderNotFound) {
			return err
		}
	}

	order := model.Order{
		Number:     orderNum,
		Status:     model.StatusNew,
		UploadedAt: time.Now(),
		UserLogin:  userLogin,
	}

	om.processingOrdersChan <- order

	return om.repo.AddOrder(userLogin, order)
}

func validateOrderNumFormat(orderNum string) bool {
	return true // TODO: алгоритм Луна
}

func (om *OrdersManager) fillProcessingChanFromRepo() {
	orders, err := om.repo.GetUnfinishedOrders()
	if err != nil {
		// TODO: логировать ошибку
		return
	}

	for _, order := range orders {
		om.processingOrdersChan <- order
	}
}

func (om *OrdersManager) checkOrderWorker() {
	for order := range om.processingOrdersChan {
		go om.processOrder(order)
	}
}

func (om *OrdersManager) processOrder(order model.Order) {
	for {
		// Запрашиваем информацию о заказе
		receivedOrder, err := om.infoGetter.GetOrderInfo(order.Number)
		if err != nil {
			if errors.Is(err, loyalsys.ErrRequestLimitReached) {
				continue
			}
			break
		}

		// Если статус заказа изменился, надо обновить значение в БД
		if receivedOrder.Status != order.Status {
			order, err = updateOrderInfo(order, receivedOrder)
			if err != nil {
				break
			}
			om.updatesChan <- order
		}

		// Если обработка завершена, завершаем функцию
		if order.Status == model.StatusProcessed || order.Status == model.StatusInvalid {
			return
		}

		time.Sleep(2 * time.Second)
	}

	// Если цикл прерван, ставим статус INVALID
	order.Status = model.StatusInvalid
	om.updatesChan <- order
}

func updateOrderInfo(order, newOrder model.Order) (model.Order, error) {
	if order.Number != newOrder.Number {
		return model.Order{}, errOrdersHaveDifferentNumbers
	}

	order.Status = newOrder.Status
	order.Accrual = newOrder.Accrual

	return order, nil
}

func (om *OrdersManager) updateRepoWorker() {
	ticker := time.NewTicker(5 * time.Second)

	var orders []model.Order

	for {
		select {
		case ord := <-om.updatesChan:
			orders = append(orders, ord)
		case <-ticker.C:
			if len(orders) == 0 {
				continue
			}
			om.repo.UpdateOrders(orders) // TODO: логировать ошибку
			orders = nil
		}
	}
}
