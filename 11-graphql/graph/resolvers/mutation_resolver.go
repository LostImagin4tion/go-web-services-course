package resolvers

import (
	"context"
	"fmt"
	"stepikGoWebServices/graph/model"
	"strconv"
)

type mutationResolver struct{ *Resolver }

// AddToCart is the resolver for the AddToCart field.
func (r *mutationResolver) AddToCart(ctx context.Context, in *model.CartInput) ([]*model.CartItem, error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var item ItemData
	found := false
	for _, i := range r.Data.Items {
		if i.ID == in.ItemID {
			item = i
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("item not found")
	}

	totalInCarts := 0
	for _, userCart := range r.UserCarts {
		for _, cartItem := range userCart {
			if cartItem.Item.ID != nil {
				cartItemID, err := strconv.Atoi(*cartItem.Item.ID)
				if err == nil && cartItemID == in.ItemID {
					totalInCarts += cartItem.Quantity
				}
			}
		}
	}

	currentInCart := 0
	cartItems := r.UserCarts[user]
	for _, cartItem := range cartItems {
		if cartItem.Item.ID != nil {
			cartItemID, err := strconv.Atoi(*cartItem.Item.ID)
			if err == nil && cartItemID == in.ItemID {
				currentInCart = cartItem.Quantity
				break
			}
		}
	}

	newQuantity := currentInCart + in.Quantity
	if totalInCarts-currentInCart+newQuantity > item.InStock {
		return nil, fmt.Errorf("not enough quantity")
	}

	itemFound := false
	for i, cartItem := range cartItems {
		if cartItem.Item.ID != nil {
			cartItemID, err := strconv.Atoi(*cartItem.Item.ID)
			if err == nil && cartItemID == in.ItemID {
				cartItems[i].Quantity = newQuantity
				itemFound = true
				break
			}
		}
	}

	if !itemFound {
		cartItems = append(cartItems, model.CartItem{
			Quantity: in.Quantity,
			Item: &model.Item{
				ID:   stringPtr(strconv.Itoa(item.ID)),
				Name: &item.Name,
			},
		})
	}

	r.UserCarts[user] = cartItems

	var result []*model.CartItem
	for _, cartItem := range cartItems {
		result = append(result, &model.CartItem{
			Quantity: cartItem.Quantity,
			Item:     cartItem.Item,
		})
	}

	return result, nil
}

// RemoveFromCart is the resolver for the RemoveFromCart field.
func (r *mutationResolver) RemoveFromCart(ctx context.Context, in model.CartInput) ([]*model.CartItem, error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cartItems := r.UserCarts[user]
	var newCartItems []model.CartItem

	for _, cartItem := range cartItems {
		if cartItem.Item.ID != nil {
			cartItemID, err := strconv.Atoi(*cartItem.Item.ID)
			if err == nil && cartItemID == in.ItemID {
				newQuantity := cartItem.Quantity - in.Quantity
				if newQuantity > 0 {
					cartItem.Quantity = newQuantity
					newCartItems = append(newCartItems, cartItem)
				}
			} else {
				newCartItems = append(newCartItems, cartItem)
			}
		} else {
			newCartItems = append(newCartItems, cartItem)
		}
	}

	r.UserCarts[user] = newCartItems

	var result []*model.CartItem
	for _, cartItem := range newCartItems {
		result = append(result, &model.CartItem{
			Quantity: cartItem.Quantity,
			Item:     cartItem.Item,
		})
	}

	return result, nil
}
