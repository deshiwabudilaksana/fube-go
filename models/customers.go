package models

import (
	"fmt"

	"github.com/deshiwabudilaksana/fube-go/database"
	"gorm.io/gorm"
)

func GetCustomer(id string) (*Customer, error) {
	var customer Customer
	result := database.DB.First(&customer, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, result.Error
	}

	return &customer, nil
}

func GetAllCustomers() (*Customer, error) {
	var customer Customer
	result := database.DB.Where("is_removed = ?", false).Find(&customer)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("not found")
		}
		return nil, result.Error
	}

	return &customer, nil
}
