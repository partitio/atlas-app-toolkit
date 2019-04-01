package gateway

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/partitio/atlas-app-toolkit/query"
)

const (
	FilterQueryKey           = "_filter"
	SortQueryKey             = "_order_by"
	FieldsQueryKey           = "_fields"
	LimitQueryKey            = "_limit"
	OffsetQueryKey           = "_offset"
	PageTokenQueryKey        = "_page_token"
	pageInfoSizeMetaKey      = "status-page-info-size"
	pageInfoOffsetMetaKey    = "status-page-info-offset"
	pageInfoPageTokenMetaKey = "status-page-info-page_token"

	query_url = "query_url"
)

// MetadataAnnotator is a function for passing metadata to a gRPC context
// It must be mainly used as ServeMuxOption for gRPC Gateway 'ServeMux'
// See: 'WithMetadata' option.
//
// MetadataAnnotator stores request URL in gRPC metadata from incoming HTTP кequest
func MetadataAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	mdmap := make(map[string]string)
	mdmap[query_url] = req.URL.String()
	return metadata.New(mdmap)
}

// SetPagination sets page info to outgoing gRPC context.
// Deprecated: Please add `infoblox.api.PageInfo` as part of gRPC message and do not rely on outgoing gRPC context.
func SetPageInfo(ctx context.Context, p *query.PageInfo) error {
	m := make(map[string]string)

	if pt := p.GetPageToken(); pt != "" {
		m[pageInfoPageTokenMetaKey] = pt
	}

	if o := p.GetOffset(); o != 0 && p.NoMore() {
		m[pageInfoOffsetMetaKey] = "null"
	} else if o != 0 {
		m[pageInfoOffsetMetaKey] = strconv.FormatUint(uint64(o), 10)
	}

	if s := p.GetSize(); s != 0 {
		m[pageInfoSizeMetaKey] = strconv.FormatUint(uint64(s), 10)
	}

	return grpc.SetHeader(ctx, metadata.New(m))
}

func ParseQuery(req interface{}, vals url.Values) (err error) {
	// extracts "_order_by" parameters from request
	if v := vals.Get(SortQueryKey); v != "" {
		s, err := query.ParseSorting(v)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		err = SetCollectionOps(req, s)
		if err != nil {
			return err
		}
	}
	// extracts "_fields" parameters from request
	if v := vals.Get(FieldsQueryKey); v != "" {
		fs := query.ParseFieldSelection(v)
		err := SetCollectionOps(req, fs)
		if err != nil {
			return err
		}
	}

	// extracts "_filter" parameters from request
	if v := vals.Get(FilterQueryKey); v != "" {
		f, err := query.ParseFiltering(v)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		err = SetCollectionOps(req, f)
		if err != nil {
			return err
		}
	}

	// extracts "_limit", "_offset",  "_page_token" parameters from request
	var p *query.Pagination
	l := vals.Get(LimitQueryKey)
	o := vals.Get(OffsetQueryKey)
	pt := vals.Get(PageTokenQueryKey)

	p, err = query.ParsePagination(l, o, pt)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = SetCollectionOps(req, p)
	if err != nil {
		return err
	}
	return nil
}
