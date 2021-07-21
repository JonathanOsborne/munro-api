package api

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MunroListQuery struct {
	SortIndex     int64
	SortParam     string
	SortDirection string
	Page          int64
	Limit         int64
	Fields        []string
	Selector      primitive.D
	FilterString  string
	Filter        primitive.M
}

func NewListQuery(c *gin.Context) *MunroListQuery {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	limit, _ := strconv.ParseInt(c.Query("limit"), 10, 64)
	sortParam := c.Query("sort_param")
	if sortParam == "" {
		sortParam = "_id"
	}
	sortDirection := c.Query("sort_direction")

	fields := c.Query("fields")

	var fieldsArray []string

	if fields != "" {
		fieldsArray = strings.Split(fields, ",")
	} else {
		fieldsArray = make([]string, 0)
	}

	filter := c.Query("filter")

	query := &MunroListQuery{
		Page:          page,
		SortParam:     sortParam,
		SortDirection: sortDirection,
		Fields:        fieldsArray,
		FilterString:  filter,
		Limit:         limit,
	}

	query.GetSortIndex()
	query.GetSelector()
	query.GetFilter()
	return query
}

func (q *MunroListQuery) GetSortIndex() {
	switch q.SortDirection {
	case "asc":
		if q.SortParam == "_id" || q.SortParam == "name" {
			q.SortIndex = -1
		} else {
			q.SortIndex = 1
		}
	case "desc":
		if q.SortParam == "_id" || q.SortParam == "name" {
			q.SortIndex = 1
		} else {
			q.SortIndex = -1
		}
	default:
		if q.SortParam == "_id" || q.SortParam == "name" {
			q.SortIndex = 1
		} else {
			q.SortIndex = -1
		}
	}
}

func (q *MunroListQuery) GetSelector() {

	var projection primitive.D

	projection = bson.D{}

	if len(q.Fields) > 0 {
		for _, field := range q.Fields {
			projection = append(projection, primitive.E{Key: field, Value: 1})
		}
		q.Selector = projection
	}
}

func (q *MunroListQuery) GetFilter() {

	var filter primitive.M

	filter = bson.M{}

	if q.FilterString == "" {
		q.Filter = filter
		return
	}

	filterStringsArray := strings.Split(q.FilterString, ",")

	if len(filterStringsArray) > 1 {
		var filtersArray []primitive.M
		for _, filterStringFragment := range filterStringsArray {
			filtersFragment := bson.M{}
			if strings.Contains(filterStringFragment, "==") {
				filtersFragment = filterEqualityOperator(filterStringFragment, "==", filtersFragment)
			} else if strings.Contains(filterStringFragment, ">=") {
				filtersFragment = filterLogicalOperator(filterStringFragment, ">=", filtersFragment)
			} else if strings.Contains(filterStringFragment, "<=") {
				filtersFragment = filterLogicalOperator(filterStringFragment, "<=", filtersFragment)
			} else if strings.Contains(filterStringFragment, ">") {
				filtersFragment = filterLogicalOperator(filterStringFragment, ">", filtersFragment)
			} else if strings.Contains(filterStringFragment, "<") {
				filtersFragment = filterLogicalOperator(filterStringFragment, "<", filtersFragment)
			} else if strings.Contains(filterStringFragment, "!=") {
				filtersFragment = filterLogicalOperator(filterStringFragment, "!=", filtersFragment)
			} else if strings.Contains(filterStringFragment, "~") {
				filter = filterContainerOperator(filterStringFragment, "~", filter)
			}
			filtersArray = append(filtersArray, filtersFragment)
		}
		filter = bson.M{"$and": filtersArray}
	} else {
		for _, filterStringFragment := range filterStringsArray {
			if strings.Contains(filterStringFragment, "==") {
				filter = filterEqualityOperator(filterStringFragment, "==", filter)
			} else if strings.Contains(filterStringFragment, ">=") {
				filter = filterLogicalOperator(filterStringFragment, ">=", filter)
			} else if strings.Contains(filterStringFragment, "<=") {
				filter = filterLogicalOperator(filterStringFragment, "<=", filter)
			} else if strings.Contains(filterStringFragment, ">") {
				filter = filterLogicalOperator(filterStringFragment, ">", filter)
			} else if strings.Contains(filterStringFragment, "<") {
				filter = filterLogicalOperator(filterStringFragment, "<", filter)
			} else if strings.Contains(filterStringFragment, "!=") {
				filter = filterLogicalOperator(filterStringFragment, "!=", filter)
			} else if strings.Contains(filterStringFragment, "~") {
				filter = filterContainerOperator(filterStringFragment, "~", filter)
			}
		}
	}
	q.Filter = filter
}

func filterEqualityOperator(filterString string, operator string, filter primitive.M) primitive.M {
	filterArray := strings.Split(filterString, operator)
	key := filterArray[0]
	value := filterArray[1]
	isNumber := isValueNumber(value)
	if isNumber {
		intValue, _ := strconv.ParseInt(value, 10, 64)
		filter[key] = intValue

	} else {
		filter[key] = value
	}
	return filter
}

func filterContainerOperator(filterString string, operator string, filter primitive.M) primitive.M {
	filterArray := strings.Split(filterString, operator)
	key := filterArray[0]
	value := filterArray[1]
	filter[key] = primitive.Regex{Pattern: value, Options: ""}

	return filter
}

func filterLogicalOperator(filterString string, operator string, filter primitive.M) primitive.M {
	filterArray := strings.Split(filterString, operator)
	key := filterArray[0]
	value := filterArray[1]
	isNumber := isValueNumber(value)

	var operatorValue string

	switch operator {
	case ">":
		operatorValue = "$gt"
	case "<":
		operatorValue = "$lt"
	case ">=":
		operatorValue = "$gte"
	case "<=":
		operatorValue = "$lte"
	case "!=":
		operatorValue = "$ne"
	}
	if isNumber {
		intValue, _ := strconv.ParseInt(value, 10, 64)
		filter[key] = bson.M{operatorValue: intValue}
	} else {
		filter[key] = bson.M{operatorValue: value}
	}

	return filter
}

func isValueNumber(test string) bool {
	matched, _ := regexp.MatchString(`^[+-]?([0-9]+\.?[0-9]*|\.[0-9]+)$`, test)
	return matched
}
