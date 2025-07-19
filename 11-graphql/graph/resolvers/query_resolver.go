package resolvers

import (
	"context"
	"fmt"
	"stepikGoWebServices/graph/model"
	"strconv"
)

type queryResolver struct{ *Resolver }

// Catalog is the resolver for the Catalog field.
func (r *queryResolver) Catalog(ctx context.Context, id *string) (*model.Catalog, error) {
	if id == nil {
		return nil, fmt.Errorf("catalog ID is required")
	}

	catalogID, err := strconv.Atoi(*id)
	if err != nil {
		return nil, err
	}

	for _, cat := range r.Data.Catalogs {
		if cat.ID == catalogID {
			catalog := &model.Catalog{
				ID:   stringPtr(strconv.Itoa(cat.ID)),
				Name: &cat.Name,
			}
			return catalog, nil
		}
	}

	return nil, fmt.Errorf("catalog not found")
}

// Shop is the resolver for the Shop field.
func (r *queryResolver) Shop(ctx context.Context, parentID *string) ([]*model.Catalog, error) {
	var results []*model.Catalog

	var targetParentID *int
	if parentID != nil {
		pid, err := strconv.Atoi(*parentID)
		if err != nil {
			return nil, err
		}
		targetParentID = &pid
	}

	for _, cat := range r.Data.Catalogs {
		if (targetParentID == nil && cat.ParentID == nil) ||
			(targetParentID != nil && cat.ParentID != nil && *cat.ParentID == *targetParentID) {
			catalog := &model.Catalog{
				ID:   stringPtr(strconv.Itoa(cat.ID)),
				Name: &cat.Name,
			}
			results = append(results, catalog)
		}
	}

	return results, nil
}

// Seller is the resolver for the Seller field.
func (r *queryResolver) Seller(ctx context.Context, id *string) (*model.Seller, error) {
	if id == nil {
		return nil, fmt.Errorf("seller ID is required")
	}

	sellerID, err := strconv.Atoi(*id)
	if err != nil {
		return nil, err
	}

	for _, seller := range r.Data.Sellers {
		if seller.ID == sellerID {
			return &model.Seller{
				ID:    stringPtr(strconv.Itoa(seller.ID)),
				Name:  &seller.Name,
				Deals: seller.Deals,
			}, nil
		}
	}

	return nil, fmt.Errorf("seller not found")
}

// MyCart is the resolver for the MyCart field.
func (r *queryResolver) MyCart(ctx context.Context) ([]*model.CartItem, error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cartItems := r.UserCarts[user]
	var result []*model.CartItem

	for _, item := range cartItems {
		result = append(result, &model.CartItem{
			Quantity: item.Quantity,
			Item:     item.Item,
		})
	}

	return result, nil
}
