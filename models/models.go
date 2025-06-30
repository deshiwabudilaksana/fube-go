package models

import (
	"time"

	"github.com/google/uuid"
)

// Role enum
type Role string

const (
	RoleOwner    Role = "OWNER"
	RoleAdmin    Role = "ADMIN"
	RoleEmployee Role = "EMPLOYEE"
	RoleCustomer Role = "CUSTOMER"
)

// User model
type User struct {
	ID         string    `gorm:"column:id;type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Username   string    `gorm:"column:username" json:"username"`
	Email      string    `gorm:"column:email" json:"email"`
	Phone      string    `gorm:"column:phone" json:"phone"`
	Role       Role      `gorm:"column:role;type:Role;default:CUSTOMER" json:"role"`
	CustomerID *int      `gorm:"column:customer_id" json:"customer_id,omitempty"`
	VendorID   string    `gorm:"column:vendor_id" json:"vendor_id"`
	Password   string    `gorm:"column:password" json:"-"` // "-" to exclude from JSON
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	IsRemoved  bool      `gorm:"column:is_removed;default:false" json:"is_removed"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// Vendor model
type Vendor struct {
	ID          string    `gorm:"column:id;type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CompanyName string    `gorm:"column:company_name" json:"company_name"`
	Email       string    `gorm:"column:email" json:"email"`
	Phone       string    `gorm:"column:phone" json:"phone"`
	Address     string    `gorm:"column:address" json:"address"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UserID      *string   `gorm:"column:user_id" json:"user_id,omitempty"`
	IsRemoved   bool      `gorm:"column:is_removed;default:false" json:"is_removed"`
}

// TableName specifies the table name for Vendor
func (Vendor) TableName() string {
	return "vendor"
}

// Customer model
type Customer struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName  string    `gorm:"column:full_name" json:"full_name"`
	Birthdate time.Time `gorm:"column:birthdate" json:"birthdate"`
	Phone     string    `gorm:"column:phone" json:"phone"`
	Points    int       `gorm:"column:points" json:"points"`
	VendorID  string    `gorm:"column:vendor_id" json:"vendor_id"`
	IsRemoved bool      `gorm:"column:is_removed;default:false" json:"is_removed"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customer"
}

// Transaction model
type Transaction struct {
	ID               string    `gorm:"column:id;type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CustomerID       int       `gorm:"column:customer_id" json:"customer_id"`
	EmployeeID       int       `gorm:"column:employee_id" json:"employee_id"`
	VendorID         string    `gorm:"column:vendor_id" json:"vendor_id"`
	PointAdded       int       `gorm:"column:point_added" json:"point_added"`
	TotalTransaction int       `gorm:"column:total_transaction" json:"total_transaction"`
	CreatedAt        time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName specifies the table name for Transaction
func (Transaction) TableName() string {
	return "transaction"
}
