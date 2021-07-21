package api

import (
	"context"
	"errors"
	"math"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Paginator struct {
	TotalRecords int64 `json:"total_records"`
	TotalPages   int64 `json:"total_pages"`
	Offset       int64 `json:"offset"`
	Limit        int64 `json:"limit"`
	Page         int64 `json:"page"`
	PrevPage     int64 `json:"prev_page"`
	NextPage     int64 `json:"next_page"`
}

type PaginationData struct {
	Total     int64 `json:"total"`
	Page      int64 `json:"page"`
	PerPage   int64 `json:"per_page"`
	Prev      int64 `json:"prev"`
	Next      int64 `json:"next"`
	TotalPages int64 `json:"total_pages"`
}

type PaginatedData struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationData `json:"pagination"`
}

type Paging struct {
	Collection  *mongo.Collection
	SortFields  bson.D
	Ctx         context.Context
	Decoder     interface{}
	Selector    interface{}
	FilterQuery interface{}
	LimitCount  int64
	PageCount   int64
}

type PagingQuery interface {
	Find() (paginatedData *PaginatedData, err error)
	Select(selector interface{}) PagingQuery
	Filter(selector interface{}) PagingQuery
	Limit(limit int64) PagingQuery
	Page(page int64) PagingQuery
	Sort(sortField string, sortValue interface{}) PagingQuery
	Decode(decode interface{}) PagingQuery
	Context(ctx context.Context) PagingQuery
}

func GeneratePaging(p *Paging, paginationInfo chan<- *Paginator) {
	var paginator Paginator
	var offset int64
	ctx := p.getContext()
	count, _ := p.Collection.CountDocuments(ctx, p.FilterQuery)

	if p.PageCount > 0 {
		offset = (p.PageCount - 1) * p.LimitCount
	} else {
		offset = 0
	}

	paginator.TotalRecords = count
	paginator.Page = p.PageCount
	paginator.Offset = offset
	paginator.Limit = p.LimitCount
	paginator.TotalPages = int64(math.Ceil(float64(count) / float64(p.LimitCount)))

	if p.PageCount > 1 {
		paginator.PrevPage = p.PageCount - 1
	} else {
		paginator.PrevPage = p.PageCount
	}

	if p.PageCount == paginator.TotalPages {
		paginator.NextPage = p.PageCount
	} else {
		paginator.NextPage = p.PageCount + 1
	}

	paginationInfo <- &paginator
}

func (p *Paginator) PaginationData() *PaginationData {
	data := PaginationData{
		Total:   p.TotalRecords,
		Page:    p.Page,
		PerPage: p.Limit, Prev: 0,
		Next:      0,
		TotalPages: p.TotalPages,
	}
	if p.Page != p.PrevPage && p.TotalRecords > 0 {
		data.Prev = p.PrevPage
	}
	if p.Page != p.NextPage && p.TotalRecords > 0 && p.Page <= p.TotalPages {
		data.Next = p.NextPage
	}

	return &data
}

func New(collection *mongo.Collection) PagingQuery {
	return &Paging{
		Collection: collection,
	}
}

func (p *Paging) Decode(decode interface{}) PagingQuery {
	p.Decoder = decode
	return p
}

func (p *Paging) Context(ctx context.Context) PagingQuery {
	p.Ctx = ctx
	return p
}

func (p *Paging) Filter(filter interface{}) PagingQuery {
	p.FilterQuery = filter
	return p
}

func (p *Paging) Select(selector interface{}) PagingQuery {
	p.Selector = selector
	return p
}

func (p *Paging) Limit(limit int64) PagingQuery {
	if limit < 1 {
		p.LimitCount = 25
	} else {
		p.LimitCount = limit
	}
	return p
}

func (p *Paging) Page(page int64) PagingQuery {
	if page < 1 {
		p.PageCount = 1
	} else {
		p.PageCount = page
	}
	return p
}

func (p *Paging) Sort(sortField string, sortValue interface{}) PagingQuery {
	sortQuery := bson.E{}
	sortQuery.Key = sortField
	sortQuery.Value = sortValue
	p.SortFields = append(p.SortFields, sortQuery)
	return p
}

func (p *Paging) Find() (data *PaginatedData, err error) {
	if p.FilterQuery == nil {
		return nil, errors.New("you need to add a filter")
	}

	paginationInfoChan := make(chan *Paginator, 1)
	GeneratePaging(p, paginationInfoChan)

	skip := getSkip(p.PageCount, p.LimitCount)

	opt := &options.FindOptions{
		Skip:  &skip,
		Limit: &p.LimitCount,
	}

	if p.Selector != nil {

		opt.SetProjection(p.Selector)
	}
	if len(p.SortFields) > 0 {
		opt.SetSort(p.SortFields)
	}

	ctx := p.getContext()

	cursor, err := p.Collection.Find(ctx, p.FilterQuery, opt)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	docs := p.Decoder

	err = cursor.All(ctx, docs)
	if err != nil {
		return nil, err
	}

	paginationInfo := <-paginationInfoChan

	result := &PaginatedData{
		Pagination: *paginationInfo.PaginationData(),
		Data:       docs,
	}

	return result, nil
}

func (p *Paging) getContext() context.Context {
	if p.Ctx != nil {
		return p.Ctx
	} else {
		return context.Background()
	}
}

func getSkip(page, limit int64) int64 {
	page--

	skip := page * limit

	if skip <= 0 {
		skip = 0
	}

	return skip
}
