package helpers

import (
	"fmt"
	"strings"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// FilterStringToQuery takes a filter search argument and converts it to a sql query
//
// Filter search arguments start with the link operator (and/or) then lists sets of column/operator/values separated by `~`
//
// eg: `filter=or~email~contains~admin~first_name~startsWith~j`
func FilterStringToQuery(filter *string) *qm.QueryMod {
	if filter == nil || len(*filter) < 4 {
		return nil
	}

	filterQueries := []string{}
	var filterValues []interface{}

	filters := strings.Split(*filter, "~")
	link := " AND "
	if filters[0] == "or" {
		link = " OR "
	}

	for i := 1; i < len(filters); i += 3 {
		column := filters[i]
		op := filters[i+1]
		value := filters[i+2]
		if value == "" {
			continue
		}

		switch op {
		case "equals":
			filterQueries = append(filterQueries, column+" = ?")
			filterValues = append(filterValues, value)
		case "contains":
			filterQueries = append(filterQueries, column+" ILIKE ?")
			filterValues = append(filterValues, "%"+value+"%")
		case "startsWith":
			filterQueries = append(filterQueries, column+" ILIKE ? ")
			filterValues = append(filterValues, value+"%")
		case "endsWith":
			filterQueries = append(filterQueries, column+" ILIKE ? ")
			filterValues = append(filterValues, "%"+value)
		}
	}
	if len(filterQueries) == 0 {
		return nil
	}

	query := qm.Where(
		"("+strings.Join(filterQueries, link)+")",
		filterValues...,
	)
	return &query
}

// SearchQuery generates a sql query for searching multiple columns
func SearchQuery(search *string, columns ...string) *qm.QueryMod {
	if search == nil {
		return nil
	}

	text := strings.ToLower(strings.Trim(*search, " "))
	if len(text) == 0 {
		return nil
	}

	queries := []string{}
	var searchText []interface{}
	for _, c := range columns {
		queries = append(queries, fmt.Sprintf("%s ILIKE ?", c))
		searchText = append(searchText, "%"+text+"%")
	}

	query := qm.Where(
		"("+strings.Join(queries, " OR ")+")",
		searchText...,
	)

	return &query
}
