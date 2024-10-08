package models

import (
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
	m "github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/middleware"
	"net/http"
)

type QueryListMetaEvent struct {
	SearchParams
	Pageable
}

type QueryListMetaEventResponse struct {
	*datastore.Filter
}

func (ql *QueryListMetaEvent) Transform(r *http.Request) (*QueryListMetaEventResponse, error) {
	searchParams, err := getSearchParams(r)
	if err != nil {
		return nil, err
	}

	return &QueryListMetaEventResponse{
		Filter: &datastore.Filter{
			SearchParams: searchParams,
			Pageable:     m.GetPageableFromContext(r.Context()),
		},
	}, nil
}

type MetaEventResponse struct {
	*datastore.MetaEvent
}
