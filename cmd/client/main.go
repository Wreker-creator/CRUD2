package main

import (
	"context"
	"log"
	"rest/protobuf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := protobuf.NewMarketServiceClient(conn)

	// demonstrate full CRUD cycle
	listAllFoodItems(client)
	addFoodItem(client, &protobuf.FoodItem{Name: "Pineapple", Price: 1.99, Calories: 100, Sugar: 0.1})
	listFoodItem(client, "Pineapple")
	updateFoodItem(client, "Pineapple", &protobuf.FoodItem{Name: "Pineapple", Price: 1.5, Calories: 100, Sugar: 0.1})
	deleteFoodItem(client, "Pineapple")
	listAllFoodItems(client) // verify Pineapple is gone
}

func listAllFoodItems(client protobuf.MarketServiceClient) {
	resp, err := client.ListAllFoodItems(context.TODO(), &protobuf.ListAllFoodItemsRequest{})
	if err != nil {
		logError(err)
	}

	for _, item := range resp.Items {
		log.Printf("Name: %v | Price: %.2f | Calories: %d | Sugar: %.2f\n",
			item.Name, item.Price, item.Calories, item.Sugar)
	}
}

// name and item are passed in so the function isn't tied to specific data
func addFoodItem(client protobuf.MarketServiceClient, item *protobuf.FoodItem) {
	resp, err := client.AddFoodItem(context.TODO(), &protobuf.AddFoodItemRequest{Item: item})
	if err != nil {
		logError(err)
	}
	log.Printf("Message: %v", resp.Message)
}

func updateFoodItem(client protobuf.MarketServiceClient, name string, item *protobuf.FoodItem) {
	resp, err := client.UpdateFoodItem(context.TODO(), &protobuf.UpdateFoodItemRequest{Name: name, Item: item})
	if err != nil {
		logError(err)
	}
	log.Printf("Message: %v", resp.Message)
}

func deleteFoodItem(client protobuf.MarketServiceClient, name string) {
	resp, err := client.DeleteFoodItem(context.TODO(), &protobuf.DeleteFoodItemRequest{Name: name})
	if err != nil {
		logError(err)
	}
	log.Printf("Message: %v", resp.Message)
}

func listFoodItem(client protobuf.MarketServiceClient, name string) {
	resp, err := client.ListFoodItem(context.TODO(), &protobuf.ListFoodItemRequest{Name: name})
	if err != nil {
		logError(err)
	}
	log.Printf("Name: %v | Price: %.2f | Calories: %d | Sugar: %.2f\n",
		resp.Item.Name, resp.Item.Price, resp.Item.Calories, resp.Item.Sugar)
}

// logError fatals on any error — fine for a demo client
// in production you'd handle each error individually
func logError(err error) {
	log.Fatalf("Failed to execute function : %v", err)
}
