package wb

import (
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
)

// Client — структура для работы с WB API
type Client struct {
	apiKey string
	client *resty.Client
}

// NewClient — создание нового клиента WB
func NewClient() *Client {
	apiKey := os.Getenv("WB_API_KEY")
	if apiKey == "" {
		panic("WB_API_KEY не установлен в переменных окружения")
	}

	client := resty.New().
		SetHostURL("https://supplies-api.wildberries.ru").
		SetHeader("Authorization", apiKey).
		SetHeader("Content-Type", "application/json")

	return &Client{
		apiKey: apiKey,
		client: client,
	}
}

// Warehouse — структура склада
type Warehouse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetWarehouses — получение списка складов
func (c *Client) GetWarehouses() ([]Warehouse, error) {
	var warehouses []Warehouse

	resp, err := c.client.R().
		SetResult(&warehouses).
		Get("/api/v1/warehouses")

	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("ошибка ответа: %s", resp.Status())
	}

	return warehouses, nil
}

// Coefficient — структура коэффициента лимита склада
type Coefficient struct {
	Date            string `json:"date"`
	Coefficient     int    `json:"coefficient"`
	WarehouseID     int    `json:"warehouseID"`
	WarehouseName   string `json:"warehouseName"`
	AllowUnload     bool   `json:"allowUnload"`
	BoxTypeName     string `json:"boxTypeName"`
	BoxTypeID       int    `json:"boxTypeID"`
	IsSortingCenter bool   `json:"isSortingCenter"`
}

// GetAcceptanceCoefficients — получение коэффициентов приёмки
func (c *Client) GetAcceptanceCoefficients() ([]Coefficient, error) {
	var coefficients []Coefficient

	resp, err := c.client.R().
		SetResult(&coefficients).
		Get("/api/v1/acceptance/coefficients")

	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("ошибка ответа: %s", resp.Status())
	}

	return coefficients, nil
}
