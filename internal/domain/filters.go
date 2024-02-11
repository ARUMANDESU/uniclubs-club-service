package domain

import (
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	"math"
)

type Filters struct {
	Page     int32
	PageSize int32
}

func (f Filters) Limit() int32 {
	return f.PageSize
}
func (f Filters) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}

type Metadata struct {
	CurrentPage  int32
	PageSize     int32
	FirstPage    int32
	LastPage     int32
	TotalRecords int32
}

func CalculateMetadata(totalRecords, page, pageSize int32) Metadata {
	if totalRecords == 0 {
		// Note that we return an empty Metadata struct if there are no records.
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int32(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

func ToPagination(m *Metadata) *clubv1.PaginationMetadata {
	return &clubv1.PaginationMetadata{
		CurrentPage:  m.CurrentPage,
		PageSize:     m.PageSize,
		FirstPage:    m.FirstPage,
		LastPage:     m.LastPage,
		TotalRecords: m.TotalRecords,
	}
}
