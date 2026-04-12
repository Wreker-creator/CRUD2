package main

import (
	"errors"
	"fmt"
	"os"
	"rest/rest"

	"github.com/spf13/cobra"
)

var store *rest.PostgresFoodStore

var rootCmd = &cobra.Command{
	Use:   "market",
	Short: "A cli tool for CRUD operations with food items",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			return errors.New("DATABASE_URL env variable is not set")
		}

		var err error
		store, err = rest.NewPostgresFoodStore(dsn)
		if err != nil {
			return fmt.Errorf("could not connect to database: %w", err)
		}

		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all items in the food market",
	RunE: func(cmd *cobra.Command, args []string) error {
		foodItems, err := store.ListAllFoodItems()
		if err != nil {
			return err
		}

		for _, item := range foodItems {
			fmt.Printf("Name: %v | Price: %.2f | Calories: %d | Sugar: %.2f\n",
				item.Name, item.Price, item.Calories, item.Sugar)
		}

		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a food item from the market",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		foodItem, err := store.ListFoodItem(name)
		if err != nil {
			return err
		}

		fmt.Printf("Name: %v | Price: %.2f | Calories: %d | Sugar: %.2f\n",
			foodItem.Name, foodItem.Price, foodItem.Calories, foodItem.Sugar)

		return nil
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a food item to the market",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		price, _ := cmd.Flags().GetFloat32("price")
		calories, _ := cmd.Flags().GetInt("calories")
		sugar, _ := cmd.Flags().GetFloat32("sugar")

		foodItem := rest.FoodItem{
			Name:     name,
			Price:    price,
			Calories: calories,
			Sugar:    sugar,
		}

		if err := store.AddFoodItem(foodItem); err != nil {
			return err
		}

		fmt.Printf("Item added successfully: %v\n", foodItem.Name)

		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the values of a food item",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		price, _ := cmd.Flags().GetFloat32("price")
		calories, _ := cmd.Flags().GetInt("calories")
		sugar, _ := cmd.Flags().GetFloat32("sugar")

		foodItem := rest.FoodItem{
			Name:     name,
			Price:    price,
			Calories: calories,
			Sugar:    sugar,
		}

		updated, err := store.UpdateFoodItem(name, foodItem)
		if err != nil {
			return err
		}
		if !updated {
			return fmt.Errorf("item with name %q does not exist", name)
		}

		// fetch updated values to confirm back to the user
		foodItem, err = store.ListFoodItem(name)
		if err != nil {
			return err
		}

		fmt.Printf("Item updated - Name: %v | Price: %.2f | Calories: %d | Sugar: %.2f\n",
			foodItem.Name, foodItem.Price, foodItem.Calories, foodItem.Sugar)

		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a food item from the market",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		deleted, err := store.DeleteFoodItem(name)
		if err != nil {
			return err
		}
		if !deleted {
			return fmt.Errorf("item with name %q does not exist", name)
		}

		fmt.Printf("Item %q deleted successfully\n", name)

		return nil
	},
}

func main() {
	// register flags on each command that needs them
	getCmd.Flags().String("name", "", "Name of the food item")

	addCmd.Flags().String("name", "", "Name of the food item")
	addCmd.Flags().Float32("price", 0, "Price of the food item")
	addCmd.Flags().Int("calories", 0, "Calories of the food item")
	addCmd.Flags().Float32("sugar", 0, "Sugar content of the food item")

	updateCmd.Flags().String("name", "", "Name of the food item")
	updateCmd.Flags().Float32("price", 0, "Price of the food item")
	updateCmd.Flags().Int("calories", 0, "Calories of the food item")
	updateCmd.Flags().Float32("sugar", 0, "Sugar content of the food item")

	deleteCmd.Flags().String("name", "", "Name of the food item")

	rootCmd.AddCommand(listCmd, getCmd, addCmd, updateCmd, deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
