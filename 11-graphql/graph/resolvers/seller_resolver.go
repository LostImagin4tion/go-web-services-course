package resolvers

import (
	"context"
	"stepikGoWebServices/graph/model"
	"strconv"
)

type sellerResolver struct{ *Resolver }

// Items is the resolver for the items field.
func (r *sellerResolver) Items(ctx context.Context, obj *model.Seller, limit *int, offset *int) ([]*model.Item, error) {
	if obj.ID == nil {
		return nil, nil
	}

	sellerID, err := strconv.Atoi(*obj.ID)
	if err != nil {
		return nil, err
	}

	var items []*model.Item
	for _, item := range r.Data.Items {
		if item.SellerID == sellerID {
			items = append(items, &model.Item{
				ID:   stringPtr(strconv.Itoa(item.ID)),
				Name: &item.Name,
			})
		}
	}

	start := 0
	if offset != nil {
		start = *offset
	}

	end := len(items)
	if limit != nil {
		end = start + *limit
		if end > len(items) {
			end = len(items)
		}
	}

	if start >= len(items) {
		return []*model.Item{}, nil
	}

	return items[start:end], nil
}
