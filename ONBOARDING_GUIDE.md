# 📖 Project FUBE: Onboarding & Product Guide

Welcome to the **Food & Beverage Unified Engine (FUBE)**. This guide will walk you through the "Little-by-Little" migration process and help you master the tools needed to optimize your restaurant's profitability.

---

## 1. Getting Started: The Onboarding Flow

### **Step 1: Create Your Account**
1.  Navigate to the **Registration** page.
2.  Enter your details and your **Company Name**.
3.  Upon registration, FUBE automatically creates a dedicated **Vendor ID** and **Main Store** for you. All your data is strictly isolated and secure.

### **Step 2: The Migration Bridge (Importing Data)**
Don't waste hours manually entering menu items. Use our **CSV Importer**.
1.  Export your "Item List" or "Sales Report" from your current POS (Square, Toast, Clover, etc.) as a CSV.
2.  Go to the **Import** tab in FUBE.
3.  Upload your CSV. FUBE's intelligent mapping engine will automatically identify your item names, prices, and external POS IDs.

---

## 2. Setting Up Your "Intelligence" Layer

### **Step 3: Ingredient & Recipe (Yield) Setup**
To track costs, FUBE needs to know what goes into your food.
1.  **Materials:** Go to the Inventory section and add your raw materials (e.g., Wagyu Beef, Brioche Bun, Sea Salt) with their current purchase prices.
2.  **Yields (Recipes):** Create "Yields" for prepped items. For example, a "Burger Prep" yield might include the beef, bun, and condiments.

### **Step 4: The Mapping UI**
This is where FUBE connects to your legacy POS.
1.  Go to the **Mapping** tab.
2.  You will see a list of items imported from your legacy POS.
3.  Click **"Link Recipe"** for each item and select the corresponding FUBE Yield/Recipe.
4.  **Result:** FUBE now knows exactly how much it costs you to serve that specific legacy item.

---

## 3. Daily Operations

### **Step 5: Yield Planning**
Before the lunch rush, use the **Planning** dashboard.
1.  Enter the number of items you **plan to sell** today.
2.  FUBE instantly calculates exactly how many grams/liters of each raw material you need to prep or order.
3.  **Pro Tip:** Click "Generate Purchase List" to see your shopping requirements.

### **Step 6: The Kitchen Display System (KDS)**
Replace paper slips with our digital board.
1.  Open the **KDS** tab on a tablet in your kitchen.
2.  New orders appear automatically (refreshed every 5 seconds).
3.  **Color Codes:**
    *   **Gray/Indigo:** New/Preparing.
    *   **Red:** Warning! Order is over 15 minutes old.
4.  **Action:** Chefs click "Start Preparing" and then "Order Ready" to move the ticket.

---

## 4. Inventory & Profit Control

### **Step 7: The Inventory Audit**
At the end of the week, check your "leaks."
1.  Go to the **Inventory** dashboard.
2.  **Theoretical Stock:** This is what FUBE calculates you *should* have based on your recipes and completed KDS orders.
3.  **Physical Stock:** Enter your actual count (e.g., "I counted 10kg of beef").
4.  **Variance:** Look for **Red** numbers. High negative variance indicates waste, theft, or over-portioning.

### **Step 8: Exporting Back to POS**
Keep your financial reports accurate.
1.  Go to the **Mapping** dashboard.
2.  Click **"Export Costs to CSV"**.
3.  Upload this file back to your main POS system to update your "Cost of Goods Sold" (COGS) and profit margin reporting.

---

## 5. Support & Troubleshooting
*   **Need a reset?** If you are in a demo environment, run the seeder script to restore sample data.
*   **Missing Items?** Ensure your CSV headers include "Name," "Price," and "SKU/ID."
*   **Offline Mode:** FUBE is designed to work on local networks. If your internet goes down, your KDS will continue to function as long as your local server is powered on.

---
*FUBE Operations Inc. - High-Performance F&B Intelligence*
