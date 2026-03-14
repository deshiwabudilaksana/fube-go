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
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
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
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
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
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
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

// FoodMaterial model
type FoodMaterial struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	VendorID  int       `gorm:"column:vendor_id" json:"vendor_id"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	UnitCost  int       `gorm:"column:unit_cost" json:"unit_cost"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Menu model
type Menu struct {
	ID            int         `gorm:"primaryKey" json:"id"`
	VendorID      int         `gorm:"column:vendor_id" json:"vendor_id"`
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	Unit          string      `json:"unit"`
	Price         int         `json:"price"`
	ExternalPosID *string     `gorm:"column:external_pos_id" json:"external_pos_id"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	MenuYields    []MenuYield `gorm:"foreignKey:MenuID" json:"menu_yields"`
}

// Yield model
type Yield struct {
	ID             int             `gorm:"primaryKey" json:"id"`
	VendorID       int             `gorm:"column:vendor_id" json:"vendor_id"`
	Name           string          `json:"name"`
	Unit           string          `json:"unit"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	YieldMaterials []YieldMaterial `gorm:"foreignKey:YieldID" json:"yield_materials"`
}

// YieldMaterial model
type YieldMaterial struct {
	ID             int          `gorm:"primaryKey" json:"id"`
	FoodMaterialID int          `gorm:"column:food_material_id" json:"food_material_id"`
	FoodMaterial   FoodMaterial `gorm:"foreignKey:FoodMaterialID" json:"food_material"`
	YieldID        int          `gorm:"column:yield_id" json:"yield_id"`
	MaterialAmount int          `gorm:"column:material_amount" json:"material_amount"`
	Unit           string       `json:"unit"`
}

// MenuYield model
type MenuYield struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	MenuID      int       `gorm:"column:menu_id" json:"menu_id"`
	YieldID     int       `gorm:"column:yield_id" json:"yield_id"`
	Yield       Yield     `gorm:"foreignKey:YieldID" json:"yield"`
	YieldAmount int       `gorm:"column:yield_amount" json:"yield_amount"`
	Unit        string    `json:"unit"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for FoodMaterial
func (FoodMaterial) TableName() string {
	return "FoodMaterial"
}

// TableName specifies the table name for Menu
func (Menu) TableName() string {
	return "Menu"
}

// TableName specifies the table name for Yield
func (Yield) TableName() string {
	return "Yield"
}

// TableName specifies the table name for YieldMaterial
func (YieldMaterial) TableName() string {
	return "YieldMaterial"
}

// TableName specifies the table name for MenuYield
func (MenuYield) TableName() string {
	return "MenuYield"
}
