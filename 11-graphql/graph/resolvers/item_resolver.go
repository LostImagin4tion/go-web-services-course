package resolvers

import (
	"context"
	"stepikGoWebServices/graph/model"
	"strconv"
)

type itemResolver struct{ *Resolver }

// Parent is the resolver for the parent field.
func (r *itemResolver) Parent(ctx context.Context, obj *model.Item) (*model.Catalog, error) {
	if obj.ID == nil {
		return nil, nil
	}

	itemID, err := strconv.Atoi(*obj.ID)
	if err != nil {
		return nil, err
	}

	for _, item := range r.Data.Items {
		if item.ID == itemID {
			for _, cat := range r.Data.Catalogs {
				if cat.ID == item.CatalogID {
					return &model.Catalog{
						ID:   stringPtr(strconv.Itoa(cat.ID)),
						Name: &cat.Name,
					}, nil
				}
			}
		}
	}

	return nil, nil
}

// Seller is the resolver for the seller field.
func (r *itemResolver) Seller(ctx context.Context, obj *model.Item) (*model.Seller, error) {
	if obj.ID == nil {
		return nil, nil
	}

	itemID, err := strconv.Atoi(*obj.ID)
	if err != nil {
		return nil, err
	}

	for _, item := range r.Data.Items {
		if item.ID == itemID {
			for _, seller := range r.Data.Sellers {
				if seller.ID == item.SellerID {
					return &model.Seller{
						ID:    stringPtr(strconv.Itoa(seller.ID)),
						Name:  &seller.Name,
						Deals: seller.Deals,
					}, nil
				}
			}
		}
	}

	return nil, nil
}

// InCart is the resolver for the inCart field.
func (r *itemResolver) InCart(ctx context.Context, obj *model.Item) (int, error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return 0, err
	}

	if obj.ID == nil {
		return 0, nil
	}

	itemID, err := strconv.Atoi(*obj.ID)
	if err != nil {
		return 0, err
	}

	cartItems := r.UserCarts[user]
	for _, item := range cartItems {
		if item.Item.ID != nil {
			cartItemID, err := strconv.Atoi(*item.Item.ID)
			if err == nil && cartItemID == itemID {
				return item.Quantity, nil
			}
		}
	}

	return 0, nil
}

// InStockText is the resolver for the inStockText field.
func (r *itemResolver) InStockText(ctx context.Context, obj *model.Item) (string, error) {
	if obj.ID == nil {
		return "", nil
	}

	itemID, err := strconv.Atoi(*obj.ID)
	if err != nil {
		return "", err
	}

	var item ItemData
	found := false
	for _, i := range r.Data.Items {
		if i.ID == itemID {
			item = i
			found = true
			break
		}
	}

	if !found {
		return "", nil
	}

	totalInCarts := 0
	for _, userCart := range r.UserCarts {
		for _, cartItem := range userCart {
			if cartItem.Item.ID != nil {
				cartItemID, err := strconv.Atoi(*cartItem.Item.ID)
				if err == nil && cartItemID == itemID {
					totalInCarts += cartItem.Quantity
				}
			}
		}
	}

	available := item.InStock - totalInCarts

	if available <= 1 {
		return "мало", nil
	} else if available >= 2 && available <= 3 {
		return "хватает", nil
	} else {
		return "много", nil
	}
}
