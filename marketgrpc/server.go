package marketgrpc

import (
	"context"
	"rest/protobuf"
	"rest/rest"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MarketServer struct {
	protobuf.UnimplementedMarketServiceServer
	// the reason for implementing this is because it gives all 5 methods a default implementation
	// that returns codes.unimplemented, so if we only implement 3 out of 5 methods, our server still compiled
	// the other 2 fall back ot he embedded default implementation.
	store rest.FoodStore
}

func NewMarketServer(store rest.FoodStore) *MarketServer {

	marketServer := MarketServer{
		store: store,
	}

	return &marketServer

}

func (s *MarketServer) ListAllFoodItems(ctx context.Context, req *protobuf.ListAllFoodItemsRequest) (*protobuf.ListAllFoodItemsResponse, error) {

	items := make([]*protobuf.FoodItem, 0)

	food, err := s.store.ListAllFoodItems()
	if err != nil {
		return nil, err
	}

	for _, foodItem := range food {
		items = append(items, &protobuf.FoodItem{
			Name:     foodItem.Name,
			Price:    foodItem.Price,
			Calories: int32(foodItem.Calories),
			Sugar:    foodItem.Sugar,
		})
	}

	return &protobuf.ListAllFoodItemsResponse{
		Items: items,
	}, nil

}

func (s *MarketServer) ListFoodItem(ctx context.Context, req *protobuf.ListFoodItemRequest) (*protobuf.ListFoodItemResponse, error) {

	foodItem, err := s.store.ListFoodItem(req.Name)
	if err != nil {
		return nil, err
	}

	return &protobuf.ListFoodItemResponse{
		Item: &protobuf.FoodItem{
			Name:     foodItem.Name,
			Price:    foodItem.Price,
			Calories: int32(foodItem.Calories),
			Sugar:    foodItem.Sugar,
		},
	}, nil

}

func (s *MarketServer) UpdateFoodItem(ctx context.Context, req *protobuf.UpdateFoodItemRequest) (*protobuf.UpdateFoodItemResponse, error) {

	if req.Item == nil {
		return nil, status.Errorf(codes.InvalidArgument, "item cannot be nil")
	}

	foodItem := rest.FoodItem{
		Name:     req.Item.Name,
		Price:    float32(req.Item.Price),
		Calories: int(req.Item.Calories),
		Sugar:    float32(req.Item.Sugar),
	}

	updated, err := s.store.UpdateFoodItem(req.Name, foodItem)
	if err != nil {
		return nil, err
	}

	if !updated {
		return nil, status.Errorf(codes.NotFound, "Failed to update, item not found %s", req.Name)
	}

	return &protobuf.UpdateFoodItemResponse{
		Message: "Successfully updated",
	}, nil

}

func (s *MarketServer) AddFoodItem(ctx context.Context, req *protobuf.AddFoodItemRequest) (*protobuf.AddFoodItemResponse, error) {

	if req.Item == nil {
		return nil, status.Errorf(codes.InvalidArgument, "item cannot be nil")
	}

	foodItem := rest.FoodItem{
		Name:     req.Item.Name,
		Price:    float32(req.Item.Price),
		Calories: int(req.Item.Calories),
		Sugar:    float32(req.Item.Sugar),
	}

	err := s.store.AddFoodItem(foodItem)
	if err != nil {
		return nil, err
	}

	return &protobuf.AddFoodItemResponse{
		Message: "Successfully added",
	}, nil

}

func (s *MarketServer) DeleteFoodItem(ctx context.Context, req *protobuf.DeleteFoodItemRequest) (*protobuf.DeleteFoodItemResponse, error) {

	deleted, err := s.store.DeleteFoodItem(req.Name)
	if err != nil {
		return nil, err
	}

	if !deleted {
		return nil, status.Errorf(codes.NotFound, "Failed to delete, item not found %s", req.Name)
	}

	return &protobuf.DeleteFoodItemResponse{
		Message: "Successfully deleted",
	}, nil

}
