package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MaterialDoc represents a raw ingredient in the MongoDB document store.
// It stores the pricing and unit information for each ingredient.
type MaterialDoc struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	VendorID          string              `bson:"vendor_id" json:"vendor_id"`
	StoreID           string              `bson:"store_id" json:"store_id"`
	Name              string              `bson:"name" json:"name"`
	Price             float64             `bson:"price" json:"price"`
	Unit              string              `bson:"unit" json:"unit"`
	DefaultSupplierID *primitive.ObjectID `bson:"default_supplier_id,omitempty" json:"default_supplier_id,omitempty"`
	CreatedAt         time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time           `bson:"updated_at" json:"updated_at"`
}

// YieldMaterialDoc represents a specific ingredient and its quantity required
// for a recipe (Yield). It is intended to be embedded within YieldDoc.
type YieldMaterialDoc struct {
	MaterialID primitive.ObjectID `bson:"material_id" json:"material_id"`
	Amount     float64            `bson:"amount" json:"amount"`
	Unit       string             `bson:"unit" json:"unit"`
}

// YieldDoc represents a recipe document in MongoDB.
// Instead of relational joins, it embeds a list of YieldMaterialDoc to optimize
// read performance for recipe compositions.
type YieldDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VendorID  string             `bson:"vendor_id" json:"vendor_id"`
	StoreID   string             `bson:"store_id" json:"store_id"`
	Name      string             `bson:"name" json:"name"`
	Unit      string             `bson:"unit" json:"unit"`
	Materials []YieldMaterialDoc `bson:"materials" json:"materials"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// MenuYieldDoc represents a reference to a recipe (Yield) and the quantity
// needed for a specific menu item.
type MenuYieldDoc struct {
	YieldID primitive.ObjectID `bson:"yield_id" json:"yield_id"`
	Amount  float64            `bson:"amount" json:"amount"`
	Unit    string             `bson:"unit" json:"unit"`
}

// MenuDoc represents a sellable item in the MongoDB document store.
// It includes metadata for external POS systems and embeds its recipe
// requirements through MenuYieldDoc structures.
type MenuDoc struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	VendorID      string                 `bson:"vendor_id" json:"vendor_id"`
	StoreID       string                 `bson:"store_id" json:"store_id"`
	ExternalPosID string                 `bson:"external_pos_id" json:"external_pos_id"`
	Name          string                 `bson:"name" json:"name"`
	Price         float64                `bson:"price" json:"price"`
	Unit          string                 `bson:"unit" json:"unit"`
	Yields        []MenuYieldDoc         `bson:"yields" json:"yields"`
	Attributes    map[string]interface{} `bson:"attributes" json:"attributes"`
	CreatedAt     time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time              `bson:"updated_at" json:"updated_at"`
}

// OrderStatus defines the lifecycle states of an order.
type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusPreparing OrderStatus = "PREPARING"
	StatusReady     OrderStatus = "READY"
	StatusServed    OrderStatus = "SERVED"
	StatusCompleted OrderStatus = "COMPLETED"
	StatusCancelled OrderStatus = "CANCELLED"
)

// TaxDetailDoc provides a breakdown of taxes applied for VAT/GST compliance.
type TaxDetailDoc struct {
	Name       string  `bson:"name" json:"name"`
	Rate       float64 `bson:"rate" json:"rate"`
	Amount     float64 `bson:"amount" json:"amount"`
	TaxAccount string  `bson:"tax_account" json:"tax_account"`
}

// PaymentDoc supports split-payment logic by tracking multiple payment entries per order.
type PaymentDoc struct {
	Method string  `bson:"method" json:"method"`
	Amount float64 `bson:"amount" json:"amount"`
	RefID  string  `bson:"ref_id" json:"ref_id"`
}

// OrderLineItem represents a specific item ordered, including its selected modifiers.
type OrderLineItem struct {
	MenuID    primitive.ObjectID `bson:"menu_id" json:"menu_id"`
	Name      string             `bson:"name" json:"name"`
	Quantity  float64            `bson:"quantity" json:"quantity"`
	Price     float64            `bson:"price" json:"price"`
	Modifiers []ModifierDoc      `bson:"modifiers" json:"modifiers"`
}

// OrderDoc represents a full transaction in the POS system.
type OrderDoc struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	VendorID    string                 `bson:"vendor_id" json:"vendor_id"`
	StoreID     string                 `bson:"store_id" json:"store_id"`
	TableNumber string                 `bson:"table_number" json:"table_number"`
	Status      OrderStatus            `bson:"status" json:"status"`
	Items       []OrderLineItem        `bson:"items" json:"items"`
	Payments    []PaymentDoc           `bson:"payments" json:"payments"`
	TaxDetails  []TaxDetailDoc         `bson:"tax_details" json:"tax_details"`
	Total       float64                `bson:"total" json:"total"`
	Metadata    map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedBy   string                 `bson:"created_by" json:"created_by"`
	UpdatedBy   string                 `bson:"updated_by" json:"updated_by"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

// InventoryDoc tracks stock levels for materials, emphasizing food safety with batch and expiry tracking.
type InventoryDoc struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VendorID   string             `bson:"vendor_id" json:"vendor_id"`
	StoreID    string             `bson:"store_id" json:"store_id"`
	MaterialID primitive.ObjectID `bson:"material_id" json:"material_id"`
	Quantity   float64            `bson:"quantity" json:"quantity"`
	Batch      string             `bson:"batch" json:"batch"`
	ExpiryDate *time.Time         `bson:"expiry_date,omitempty" json:"expiry_date,omitempty"`
	CreatedBy  string             `bson:"created_by" json:"created_by"`
	UpdatedBy  string             `bson:"updated_by" json:"updated_by"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// ModifierOption defines a single choice within a modifier group.
type ModifierOption struct {
	Name  string  `bson:"name" json:"name"`
	Price float64 `bson:"price" json:"price"`
}

// ModifierGroupDoc supports menu customizations like 'Toppings' or 'Doneness'.
type ModifierGroupDoc struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VendorID string             `bson:"vendor_id" json:"vendor_id"`
	StoreID  string             `bson:"store_id" json:"store_id"`
	Name     string             `bson:"name" json:"name"`
	Options  []ModifierOption   `bson:"options" json:"options"`
}

// ModifierDoc is the snapshot of a selected modifier embedded in an OrderLineItem.
type ModifierDoc struct {
	ModifierID primitive.ObjectID `bson:"modifier_id" json:"modifier_id"`
	Name       string             `bson:"name" json:"name"`
	Price      float64            `bson:"price" json:"price"`
}

// UserDoc represents a user in MongoDB.
type UserDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"`
	Role      Role               `bson:"role" json:"role"`
	VendorID  string             `bson:"vendor_id" json:"vendor_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// VendorDoc represents a vendor (restaurant/business) in MongoDB.
type VendorDoc struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CompanyName string             `bson:"company_name" json:"company_name"`
	Email       string             `bson:"email" json:"email"`
	Phone       string             `bson:"phone" json:"phone"`
	Address     string             `bson:"address" json:"address"`
	UserID      string             `bson:"user_id" json:"user_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
