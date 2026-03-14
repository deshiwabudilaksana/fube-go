package models

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/deshiwabudilaksana/fube-go/database"
)

// MenuCostResult represents the cost breakdown for a menu item
type MenuCostResult struct {
	MenuID    string  `json:"menu_id"`
	MenuName  string  `json:"menu_name"`
	TotalCost float64 `json:"total_cost"`
}

// MaterialRequirement represents the aggregated requirement for a food material
type MaterialRequirement struct {
	MaterialID   string  `json:"material_id"`
	MaterialName string  `json:"material_name"`
	TotalAmount  float64 `json:"total_amount"`
	Unit         string  `json:"unit"`
}

// ProductionPlanInput represents the input for production planning
type ProductionPlanInput struct {
	MenuID          string `json:"menu_id"`
	PlannedQuantity int    `json:"planned_quantity"`
}

// ExternalReportData represents the data exported for external reporting
type ExternalReportData struct {
	ExternalPosID string  `json:"external_pos_id"`
	MenuName      string  `json:"menu_name"`
	TotalCost     float64 `json:"total_cost"`
	SellingPrice  float64 `json:"selling_price"`
	Yields        []struct {
		Name   string `json:"name"`
		Amount int    `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"yields"`
}

// GetMenuCost calculates the total cost of a menu item by traversing
// MenuYield -> Yield -> YieldMaterial -> FoodMaterial.
func GetMenuCost(menuID int, vendorID int) (*MenuCostResult, error) {
	var menu Menu
	// Preload all necessary relationships for cost calculation in a single query
	err := database.DB.Where("vendor_id = ?", vendorID).
		Preload("MenuYields.Yield.YieldMaterials.FoodMaterial").
		First(&menu, menuID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch menu with ID %d for vendor %d: %w", menuID, vendorID, err)
	}

	totalCost := calculateMenuCost(&menu)
	return &MenuCostResult{
		MenuID:    fmt.Sprintf("%d", menu.ID),
		MenuName:  menu.Name,
		TotalCost: float64(totalCost),
	}, nil
}

// PlanProduction calculates total aggregated amount of FoodMaterial required
// based on a list of MenuID and PlannedQuantity.
func PlanProduction(inputs []ProductionPlanInput, vendorID int) ([]MaterialRequirement, error) {
	if len(inputs) == 0 {
		return []MaterialRequirement{}, nil
	}

	menuIDs := make([]int, 0, len(inputs))
	inputMap := make(map[int]int)
	for _, input := range inputs {
		id, err := strconv.Atoi(input.MenuID)
		if err != nil {
			continue // Skip invalid IDs for the SQL version
		}
		menuIDs = append(menuIDs, id)
		inputMap[id] += input.PlannedQuantity // Aggregate quantities if duplicate menu IDs provided
	}

	var menus []Menu
	err := database.DB.Where("vendor_id = ?", vendorID).
		Preload("MenuYields.Yield.YieldMaterials.FoodMaterial").
		Where("id IN ?", menuIDs).
		Find(&menus).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch menus for planning for vendor %d: %w", vendorID, err)
	}

	requirementsMap := make(map[int]*MaterialRequirement)
	for _, menu := range menus {
		plannedQty := inputMap[menu.ID]
		for _, my := range menu.MenuYields {
			for _, ym := range my.Yield.YieldMaterials {
				fm := ym.FoodMaterial
				amount := plannedQty * my.YieldAmount * ym.MaterialAmount

				if req, ok := requirementsMap[fm.ID]; ok {
					req.TotalAmount += float64(amount)
				} else {
					requirementsMap[fm.ID] = &MaterialRequirement{
						MaterialID:   fmt.Sprintf("%d", fm.ID),
						MaterialName: fm.Name,
						TotalAmount:  float64(amount),
						Unit:         fm.Unit,
					}
				}
			}
		}
	}

	result := make([]MaterialRequirement, 0, len(requirementsMap))
	for _, req := range requirementsMap {
		result = append(result, *req)
	}

	// Sort by MaterialID for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].MaterialID < result[j].MaterialID
	})

	return result, nil
}

// GetExternalReportingData generates a JSON payload mapping Menu.external_pos_id
// to FUBE's calculated cost/yield data.
func GetExternalReportingData(vendorID int) ([]ExternalReportData, error) {
	var menus []Menu
	// Filter for menus that have an external_pos_id
	err := database.DB.Where("vendor_id = ?", vendorID).
		Preload("MenuYields.Yield.YieldMaterials.FoodMaterial").
		Where("external_pos_id IS NOT NULL").
		Find(&menus).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch menus for external reporting for vendor %d: %w", vendorID, err)
	}

	report := make([]ExternalReportData, 0, len(menus))
	for _, menu := range menus {
		if menu.ExternalPosID == nil {
			continue
		}

		totalCost := calculateMenuCost(&menu)

		data := ExternalReportData{
			ExternalPosID: *menu.ExternalPosID,
			MenuName:      menu.Name,
			TotalCost:     float64(totalCost),
			SellingPrice:  float64(menu.Price),
		}

		for _, my := range menu.MenuYields {
			data.Yields = append(data.Yields, struct {
				Name   string `json:"name"`
				Amount int    `json:"amount"`
				Unit   string `json:"unit"`
			}{
				Name:   my.Yield.Name,
				Amount: my.YieldAmount,
				Unit:   my.Unit,
			})
		}
		report = append(report, data)
	}

	return report, nil
}

// calculateMenuCost is a helper function to calculate cost from a preloaded Menu model
func calculateMenuCost(menu *Menu) int {
	totalCost := 0
	for _, my := range menu.MenuYields {
		for _, ym := range my.Yield.YieldMaterials {
			// Cost = (Yields in Menu) * (Materials in Yield) * (Cost per Material Unit)
			totalCost += my.YieldAmount * ym.MaterialAmount * ym.FoodMaterial.UnitCost
		}
	}
	return totalCost
}
