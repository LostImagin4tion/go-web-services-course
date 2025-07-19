package resolvers

import (
	"context"
	"stepikGoWebServices/graph/model"
	"strconv"
)

type catalogResolver struct{ *Resolver }

// Parent is the resolver for the parent field.
func (r *catalogResolver) Parent(ctx context.Context, obj *model.Catalog) (*model.Catalog, error) {
	if obj.ID == nil {
		return nil, nil
	}

	catalogID, err := strconv.Atoi(*obj.ID)
	if err != nil {
		return nil, err
	}

	for _, cat := range r.Data.Catalogs {
		if cat.ID == catalogID && cat.ParentID != nil {
			for _, parentCat := range r.Data.Catalogs {
				if parentCat.ID == *cat.ParentID {
					return &model.Catalog{
						ID:   stringPtr(strconv.Itoa(parentCat.ID)),
						Name: &parentCat.Name,
					}, nil
				}
			}
		}
	}

	return nil, nil
}

// Childs is the resolver for the childs field.
func (r *catalogResolver) Childs(ctx context.Context, obj *model.Catalog) ([]*model.Catalog, error) {
	if obj.ID == nil {
		return nil, nil
	}

	catalogID, err := strconv.Atoi(*obj.ID)
	if err != nil {
		return nil, err
	}

	var children []*model.Catalog
	for _, cat := range r.Data.Catalogs {
		if cat.ParentID != nil && *cat.ParentID == catalogID {
			children = append(children, &model.Catalog{
				ID:   stringPtr(strconv.Itoa(cat.ID)),
				Name: &cat.Name,
			})
		}
	}

	return children, nil
}

// Items is the resolver for the items field.
func (r *catalogResolver) Items(ctx context.Context, obj *model.Catalog, limit *int, offset *int) ([]*model.Item, error) {
	if obj.ID == nil {
		return nil, nil
	}

	catalogID, err := strconv.Atoi(*obj.ID)
	if err != nil {
		return nil, err
	}

	var items []*model.Item
	for _, item := range r.Data.Items {
		if item.CatalogID == catalogID {
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
