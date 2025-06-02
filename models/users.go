package models

import (
	"fmt"
	"log"

	"github.com/deshiwabudilaksana/fube-go/database"
	"gorm.io/gorm"
)

func GetUser(id string) (*User, error) {
	var user User
	result := database.DB.First(&user, id) // Use GORM's First method

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

func GetAllUsers() (*User, error) {
	var users User
	result := database.DB.Where("is_removed = ?", false).Find(&users)

	log.Println(users.Password)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("not found")
		}
		return nil, result.Error
	}

	return &users, nil
}
