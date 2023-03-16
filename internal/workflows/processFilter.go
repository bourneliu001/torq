package workflows

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/tags"
)

type FilterCategoryType string

const (
	FilterCategoryTypeNumber   FilterCategoryType = "number"
	FilterCategoryTypeString   FilterCategoryType = "string"
	FilterCategoryTypeDate     FilterCategoryType = "date"
	FilterCategoryTypeBoolean  FilterCategoryType = "boolean"
	FilterCategoryTypeArray    FilterCategoryType = "array"
	FilterCategoryTypeDuration FilterCategoryType = "duration"
	FilterCategoryTypeEnum     FilterCategoryType = "enum"
	FilterCategoryTypeTag      FilterCategoryType = "tag"
)

type FilterClauses struct {
	And    []FilterClauses `json:"$and"`
	Or     []FilterClauses `json:"$or"`
	Filter Filter          `json:"$filter"`
}
type Parameter string

type Filter struct {
	FuncName  string      `json:"funcName"`
	Key       string      `json:"key"`
	Parameter interface{} `json:"parameter"`
	Category  string      `json:"category"`
}

type FilterParameterType interface{}

type FilterFunc func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool

type SelectOptionType struct {
	Value string
	Label string
}

type FilterInterface struct {
	Category      FilterCategoryType
	FuncName      string
	Parameter     FilterParameterType
	Key           string
	SelectOptions []SelectOptionType
	Value         interface{}
	Label         string
}

type FilterClause struct {
	Prefix string
	Filter FilterInterface
	Result bool
}

type AndClause struct {
	Prefix       string
	ChildClauses []interface{}
	Result       bool
}

func (f *FilterClause) ToJSON() map[string]interface{} {
	return map[string]interface{}{f.Prefix: f.Filter}
}

func (a *AndClause) AddChildClause(clause Clause) {
	a.ChildClauses = append(a.ChildClauses, clause)
}

type OrClause struct {
	Prefix       string
	ChildClauses []interface{}
	Result       bool
}

type Clause struct {
	Prefix       string
	ChildClauses []Clause
	Filter       FilterInterface
	Result       bool
}

type ClauseWithResult struct {
	Prefix       string
	ChildClauses []interface{}
	Result       bool
}

func (c *Clause) parseClause(data map[string]interface{}) {
	for i, childClause := range c.ChildClauses {

		switch childClause.Prefix {
		case "$and":
			newChildClaud := Clause{
				Prefix:       childClause.Prefix,
				ChildClauses: childClause.ChildClauses,
			}
			newChildClaud.parseClause(data)

			if newChildClaud.Result {
				c.ChildClauses[i].Result = true
			}

			if allTrue := true; allTrue {
				for _, subChildClause := range newChildClaud.ChildClauses {
					if !subChildClause.Result {
						allTrue = false
						break
					}
				}
				if allTrue {
					c.Result = true
				}
			}
		case "$or":
			newChildClaud := Clause{
				Prefix:       childClause.Prefix,
				ChildClauses: childClause.ChildClauses,
			}
			newChildClaud.parseClause(data)
			for _, subChildClause := range newChildClaud.ChildClauses {
				if subChildClause.Result {
					c.ChildClauses[i].Result = true
				}
			}
		case "$filter":
			filterClause := FilterClause{
				Prefix: childClause.Prefix,
				Filter: childClause.Filter,
			}
			filterFunc, ok := GetFilterFunctions()[filterClause.Filter.Category][filterClause.Filter.FuncName]
			if !ok {
				panic("Filter function is not yet defined")
			}
			if _, ok := data[strings.ToLower(filterClause.Filter.Key)]; ok {
				c.ChildClauses[i].Result = filterFunc(data, strings.ToLower(filterClause.Filter.Key), filterClause.Filter.Parameter)
			} else {
				c.ChildClauses[i].Result = true
			}
		}
	}
}

func GetFilterFunctions() map[FilterCategoryType]map[string]FilterFunc {
	return map[FilterCategoryType]map[string]FilterFunc{
		FilterCategoryTypeNumber: {
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat == filterValueFloat
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return true
				}
				if filterValue == nil {
					return false
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat != filterValueFloat
			},
			"gt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return false
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat > filterValueFloat
			},
			"gte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat >= filterValueFloat
			},
			"lt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return false
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat < filterValueFloat
			},
			"lte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat <= filterValueFloat
			},
		},
		FilterCategoryTypeDuration: {
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat == filterValueFloat
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return true
				}
				if filterValue == nil {
					return false
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat != filterValueFloat
			},
			"gt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return false
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat > filterValueFloat
			},
			"gte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat >= filterValueFloat
			},
			"lt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return false
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat < filterValueFloat
			},
			"lte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat <= filterValueFloat
			},
		},
		FilterCategoryTypeString: {
			"like": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueString, ok := dataMap[dataKey].(string)
				if !ok {
					dataValueStringPointer, ok := dataMap[dataKey].(*string)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: dataValueString) so defaulting to false instead of a panic!")
						return false
					}
					dataValueString = *dataValueStringPointer
				}
				filterValueString, ok := filterValue.(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: filterValueString) so defaulting to false instead of a panic!")
					return false
				}
				return strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
			},
			"notLike": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return true
				}
				if filterValue == nil {
					return false
				}
				dataValueString, ok := dataMap[dataKey].(string)
				if !ok {
					dataValueStringPointer, ok := dataMap[dataKey].(*string)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: dataValueString) so defaulting to false instead of a panic!")
						return false
					}
					dataValueString = *dataValueStringPointer
				}
				filterValueString, ok := filterValue.(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: filterValueString) so defaulting to false instead of a panic!")
					return false
				}
				return !strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
			},
		},
		FilterCategoryTypeEnum: {
			// TODO: This needs to be fixed. Enums are using any and notAny instead of like and notLike
			"like": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueString, ok := dataMap[dataKey].(string)
				if !ok {
					dataValueStringPointer, ok := dataMap[dataKey].(*string)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: dataValueString) so defaulting to false instead of a panic!")
						return false
					}
					dataValueString = *dataValueStringPointer
				}
				filterValueString, ok := filterValue.(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: filterValueString) so defaulting to false instead of a panic!")
					return false
				}
				return strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
			},
			"notLike": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return true
				}
				if filterValue == nil {
					return false
				}
				dataValueString, ok := dataMap[dataKey].(string)
				if !ok {
					dataValueStringPointer, ok := dataMap[dataKey].(*string)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: dataValueString) so defaulting to false instead of a panic!")
						return false
					}
					dataValueString = *dataValueStringPointer
				}
				filterValueString, ok := filterValue.(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: filterValueString) so defaulting to false instead of a panic!")
					return false
				}
				return !strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
			},
		},
		FilterCategoryTypeDate: {
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueTime, ok := dataMap[dataKey].(time.Time)
				if !ok {
					dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
						return false
					}
					dataValueTime = *dataValueTimePointer
				}
				dataValueTime = truncateToMinute(dataValueTime)
				filterValueTime, timeOk := filterValue.(time.Time)
				if timeOk {
					filterValueTime = truncateToMinute(filterValueTime)
				}
				filterValueString, stringOk := filterValue.(string)
				if stringOk {
					var err error
					filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
					if err != nil {
						log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
						return false
					}
				}
				if !timeOk && !stringOk {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueTime.Equal(filterValueTime)
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return true
				}
				if filterValue == nil {
					return false
				}
				dataValueTime, ok := dataMap[dataKey].(time.Time)
				if !ok {
					dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
						return false
					}
					dataValueTime = *dataValueTimePointer
				}
				dataValueTime = truncateToMinute(dataValueTime)
				filterValueTime, timeOk := filterValue.(time.Time)
				if timeOk {
					filterValueTime = truncateToMinute(filterValueTime)
				}
				filterValueString, stringOk := filterValue.(string)
				if stringOk {
					var err error
					filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
					if err != nil {
						log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
						return false
					}
				}
				if !timeOk && !stringOk {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
					return false
				}
				return !dataValueTime.Equal(filterValueTime)
			},
			"gt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return false
				}
				dataValueTime, ok := dataMap[dataKey].(time.Time)
				if !ok {
					dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
						return false
					}
					dataValueTime = *dataValueTimePointer
				}
				dataValueTime = truncateToMinute(dataValueTime)
				filterValueTime, timeOk := filterValue.(time.Time)
				if timeOk {
					filterValueTime = truncateToMinute(filterValueTime)
				}
				filterValueString, stringOk := filterValue.(string)
				if stringOk {
					var err error
					filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
					if err != nil {
						log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
						return false
					}
				}
				if !timeOk && !stringOk {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueTime.After(filterValueTime)
			},
			"gte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueTime, ok := dataMap[dataKey].(time.Time)
				if !ok {
					dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
						return false
					}
					dataValueTime = *dataValueTimePointer
				}
				dataValueTime = truncateToMinute(dataValueTime)
				filterValueTime, timeOk := filterValue.(time.Time)
				if timeOk {
					filterValueTime = truncateToMinute(filterValueTime)
				}
				filterValueString, stringOk := filterValue.(string)
				if stringOk {
					var err error
					filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
					if err != nil {
						log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
						return false
					}
				}
				if !timeOk && !stringOk {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueTime.After(filterValueTime) || dataValueTime.Equal(filterValueTime)
			},
			"lt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return false
				}
				dataValueTime, ok := dataMap[dataKey].(time.Time)
				if !ok {
					dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
						return false
					}
					dataValueTime = *dataValueTimePointer
				}
				dataValueTime = truncateToMinute(dataValueTime)
				filterValueTime, timeOk := filterValue.(time.Time)
				if timeOk {
					filterValueTime = truncateToMinute(filterValueTime)
				}
				filterValueString, stringOk := filterValue.(string)
				if stringOk {
					var err error
					filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
					if err != nil {
						log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
						return false
					}
				}
				if !timeOk && !stringOk {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueTime.Before(filterValueTime)
			},
			"lte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				dataValueTime, ok := dataMap[dataKey].(time.Time)
				if !ok {
					dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
					if !ok {
						log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
						return false
					}
					dataValueTime = *dataValueTimePointer
				}
				dataValueTime = truncateToMinute(dataValueTime)
				filterValueTime, timeOk := filterValue.(time.Time)
				if timeOk {
					filterValueTime = truncateToMinute(filterValueTime)
				}
				filterValueString, stringOk := filterValue.(string)
				if stringOk {
					var err error
					filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
					if err != nil {
						log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
						return false
					}
				}
				if !timeOk && !stringOk {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueTime.Before(filterValueTime) || dataValueTime.Equal(filterValueTime)
			},
		},
		FilterCategoryTypeBoolean: {
			"eq":  filterCategoryTypeBooleanEq,
			"neq": filterCategoryTypeBooleanNeq,
		},
		FilterCategoryTypeArray: {
			// TODO: This needs to be fixed. Arrays are using any and notAny instead of like and notLike
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return false
				}
				if filterValue == nil {
					return true
				}
				// TODO FIXME this will not work and panic???
				dataValueArray, ok := dataMap[dataKey].([]interface{})
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeArray: dataValueArray) so defaulting to false!")
					return false
				}
				filterValueArray, ok := filterValue.([]interface{})
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeArray: filterValueArray) so defaulting to false!")
					return false
				}
				for _, dataValueEntry := range dataValueArray {
					foundIt := false
					for _, filterValueEntry := range filterValueArray {
						if dataValueEntry == filterValueEntry {
							foundIt = true
						}
					}
					if !foundIt {
						return false
					}
				}
				return true
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if isNil(dataMap[dataKey]) != (filterValue == nil) {
					return true
				}
				if filterValue == nil {
					return false
				}
				// TODO FIXME this will not work and panic???
				dataValueArray, ok := dataMap[dataKey].([]interface{})
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeArray: dataValueArray) so defaulting to false!")
					return false
				}
				filterValueArray, ok := filterValue.([]interface{})
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeArray: filterValueArray) so defaulting to false!")
					return false
				}
				for _, dataValueEntry := range dataValueArray {
					foundIt := false
					for _, filterValueEntry := range filterValueArray {
						if dataValueEntry == filterValueEntry {
							foundIt = true
						}
					}
					if !foundIt {
						return true
					}
				}
				return false
			},
		},
		FilterCategoryTypeTag: {
			"any":    filterCategoryTypeTagAny,
			"notAny": filterCategoryTypeTagNotAny,
		},
	}
}

func isNil(unknownType interface{}) bool {
	if unknownType == nil {
		return true
	}
	switch unknownType.(type) {
	case *float64:
		return unknownType == (*float64)(nil)
	case *float32:
		return unknownType == (*float32)(nil)
	case *int64:
		return unknownType == (*int64)(nil)
	case *int32:
		return unknownType == (*int32)(nil)
	case *int16:
		return unknownType == (*int16)(nil)
	case *int8:
		return unknownType == (*int8)(nil)
	case *int:
		return unknownType == (*int)(nil)
	case *uint64:
		return unknownType == (*uint64)(nil)
	case *uint32:
		return unknownType == (*uint32)(nil)
	case *uint16:
		return unknownType == (*uint16)(nil)
	case *uint8:
		return unknownType == (*uint8)(nil)
	case *uint:
		return unknownType == (*uint)(nil)
	case *bool:
		return unknownType == (*bool)(nil)
	case *string:
		return unknownType == (*string)(nil)
	case *time.Time:
		return unknownType == (*time.Time)(nil)
	case *interface{}:
		return unknownType == (*interface{})(nil)
	}
	return false
}

func truncateToMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
}

func getFloats(dataValueOfUnknownType interface{}, filterValueOfUnknownType interface{}) (float64, float64, error) {
	dataValueFloat, err := getFloat(dataValueOfUnknownType)
	if err != nil {
		return math.NaN(), math.NaN(), errors.Wrap(err, "converting dataValueOfUnknownType into dataValueFloat")
	}
	filterValueFloat, err := getFloat(filterValueOfUnknownType)
	if err != nil {
		return math.NaN(), math.NaN(), errors.Wrap(err, "converting filterValueOfUnknownType into filterValueFloat")
	}
	return dataValueFloat, filterValueFloat, nil
}

//func getInt(unknownType interface{}) (int64, error) {
//	switch i := unknownType.(type) {
//	case float64:
//		return int64(i), nil
//	case float32:
//		return int64(i), nil
//	case int64:
//		return i, nil
//	case int32:
//		return int64(i), nil
//	case int:
//		return int64(i), nil
//	case uint64:
//		return int64(i), nil
//	case uint32:
//		return int64(i), nil
//	case uint:
//		return int64(i), nil
//	case string:
//		parameter, err := strconv.ParseInt(i, 10, 64)
//		if err != nil {
//			log.Debug().Err(err).Msgf("Failed to convert string to int64 while filtering")
//		}
//		return parameter, err
//	default:
//		var floatType = reflect.TypeOf(int64(0))
//		var stringType = reflect.TypeOf("")
//		v := reflect.ValueOf(unknownType)
//		v = reflect.Indirect(v)
//		if v.Type().ConvertibleTo(floatType) {
//			fv := v.Convert(floatType)
//			return fv.Int(), nil
//		} else if v.Type().ConvertibleTo(stringType) {
//			sv := v.Convert(stringType)
//			s := sv.String()
//			parameter, err := strconv.ParseInt(s, 10, 64)
//			if err != nil {
//				log.Debug().Err(err).Msgf("Failed to convert string to int64 while filtering")
//			}
//			return parameter, err
//		} else {
//			return 0, fmt.Errorf("can't convert %v to float64", v.Type())
//		}
//	}
//}

// Note that large int64's (over 2 pow 53) will be rounded when converted to float64,
// so if you're thinking of float64 as a "universal" number type, take its limitations into account.
func getFloat(unknownType interface{}) (float64, error) {
	switch i := unknownType.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		parameter, err := strconv.ParseFloat(i, 64)
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to convert string to float64 while filtering")
		}
		return parameter, errors.Wrap(err, "converting string to float64 while filtering")
	default:
		var floatType = reflect.TypeOf(float64(0))
		var stringType = reflect.TypeOf("")
		v := reflect.ValueOf(unknownType)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			parameter, err := strconv.ParseFloat(s, 64)
			if err != nil {
				log.Debug().Err(err).Msgf("Failed to convert string to float64 while filtering")
			}
			return parameter, errors.Wrap(err, "converting string to float64 while filtering")
		} else {
			return math.NaN(), errors.New(fmt.Sprintf("can't convert %v to float64", v.Type()))
		}
	}
}

func ApplyFilters(filters interface{}, data []map[string]interface{}) []interface{} {
	var result []interface{}
	for _, item := range data {
		if ProcessQuery(DeserialiseQuery(filters), item) {
			result = append(result, item)
		}
	}
	return result
}

func ProcessQuery(filters interface{}, item map[string]interface{}) bool {
	var result bool

	switch filters.(Clause).Prefix {
	case "$and":
		andFilter := Clause{
			Prefix:       filters.(Clause).Prefix,
			ChildClauses: filters.(Clause).ChildClauses,
		}
		andFilter.parseClause(item)
		result = andFilter.Result
		if allTrue := true; allTrue {
			for _, childClause := range filters.(Clause).ChildClauses {
				if !childClause.Result {
					allTrue = false
					break
				}
			}
			if allTrue {
				result = true
			}
		}
	case "$or":
		orFilter := Clause{
			Prefix:       filters.(Clause).Prefix,
			ChildClauses: filters.(Clause).ChildClauses,
		}
		orFilter.parseClause(item)
		result = orFilter.Result
		for _, childClause := range filters.(Clause).ChildClauses {
			if childClause.Result {
				result = childClause.Result
				break
			}
		}
	case "$filter":
		filterClause := Clause{
			Prefix: filters.(Clause).Prefix,
			Filter: filters.(Clause).Filter,
		}
		filterClause.parseClause(item)
		result = filterClause.Result
	}
	return result
}

func DeserialiseQuery(query interface{}) interface{} {
	if query == nil {
		return Clause{}
	}
	if query.(FilterClauses).Filter.FuncName != "" {

		filter := FilterInterface{
			Parameter: query.(FilterClauses).Filter.Parameter,
			FuncName:  query.(FilterClauses).Filter.FuncName,
			Key:       query.(FilterClauses).Filter.Key,
			Category:  FilterCategoryType(query.(FilterClauses).Filter.Category),
		}
		return Clause{
			Prefix: "$filter",
			Filter: filter,
		}
	}
	if query.(FilterClauses).And != nil {
		var subClauses []Clause
		for _, subclause := range query.(FilterClauses).And {
			theClause := DeserialiseQuery(subclause)
			subClauses = append(subClauses, theClause.(Clause))
		}
		return Clause{
			Prefix:       "$and",
			ChildClauses: subClauses,
		}
	}
	if query.(FilterClauses).Or != nil {
		var subClauses []Clause
		for _, subclause := range query.(FilterClauses).Or {
			theClause := DeserialiseQuery(subclause)
			subClauses = append(subClauses, theClause.(Clause))
		}
		return Clause{
			Prefix:       "$or",
			ChildClauses: subClauses}
	}
	panic("Expected JSON to contain $filter, $or or $and")
}

func filterCategoryArrayAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueArray, ok := dataMap[dataKey].([]interface{})
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryArray: dataValueArray) so defaulting to false!")
		return false
	}
	filterValueArray, ok := filterValue.([]interface{})
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryArray: filterValueArray) so defaulting to false!")
		return false
	}
	for _, filterValueItem := range filterValueArray {
		for _, dataValueItem := range dataValueArray {
			if filterValueItem == dataValueItem {
				return true
			}
		}
	}
	return false
}

func filterCategoryArrayNotAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueArray, ok := dataMap[dataKey].([]interface{})
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryArray: dataValueArray) so defaulting to false!")
		return false
	}
	filterValueArray, ok := filterValue.([]interface{})
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryArray: filterValueArray) so defaulting to false!")
		return false
	}
	for _, filterValueItem := range filterValueArray {
		for _, dataValueItem := range dataValueArray {
			if filterValueItem == dataValueItem {
				return false
			}
		}
	}
	return true
}

func filterCategoryTypeTagAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return true
	}
	if filterValue == nil && dataMap[dataKey] == nil {
		return true
	}
	if filterValue == nil {
		return true
	}
	dataValueTags, ok := dataMap[dataKey].([]tags.Tag)
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeTag: dataValueTags) so defaulting to false!")
		return false
	}
	filterValueTagResponses, tagResponsesOk := filterValue.([]tags.TagResponse)
	if tagResponsesOk {
		for _, tag := range filterValueTagResponses {
			for _, dataValueTag := range dataValueTags {
				if tag.TagId == dataValueTag.TagId {
					return true
				}
			}
		}
	}
	filterValueTags, tagsOk := filterValue.([]tags.Tag)
	if tagsOk {
		for _, tag := range filterValueTags {
			for _, dataValueTag := range dataValueTags {
				if tag.TagId == dataValueTag.TagId {
					return true
				}
			}
		}
	}
	filterValueTagIdsO, tagIdsOk := filterValue.([]interface{})
	if tagIdsOk {
	filterLoop:
		for _, tagIdO := range filterValueTagIdsO {
			for _, dataValueTag := range dataValueTags {
				tagId, err := getFloat(tagIdO)
				if err != nil {
					log.Error().Msgf("could not run convert interface into tagId in filter function (FilterCategoryTypeTag)")
					tagIdsOk = false
					break filterLoop
				}
				if int(tagId) == dataValueTag.TagId {
					return true
				}
			}
		}
	}
	if !tagsOk && !tagResponsesOk && !tagIdsOk {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeTag) so defaulting to false!")
		return false
	}
	return false
}

func filterCategoryTypeTagNotAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {

	if filterValue == nil && dataMap[dataKey] == nil {
		return false
	}
	// If the filter value is nil, returns false.
	if filterValue == nil {
		return false
	}
	// Converts the data value to a slice of tags.Tag, and if it fails, returns false.
	dataValueTags, ok := dataMap[dataKey].([]tags.Tag)
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeTag: dataValueTags) so defaulting to false!")
		return false
	}

	// If the filter value is a slice of tags.TagResponse, iterates over the filter value tags and the data value tags
	//   and returns false if any of the filter tags match any of the data value tags.
	filterValueTagResponses, tagResponsesOk := filterValue.([]tags.TagResponse)
	if tagResponsesOk {
		for _, tag := range filterValueTagResponses {
			for _, dataValueTag := range dataValueTags {
				if tag.TagId == dataValueTag.TagId {
					return false
				}
			}
		}
	}

	// If the filter value is a slice of tags.Tag, iterates over the filter value tags and the data value tags
	//   and returns false if any of the filter tags match any of the data value tags.
	filterValueTags, tagsOk := filterValue.([]tags.Tag)
	if tagsOk {
		for _, tag := range filterValueTags {
			for _, dataValueTag := range dataValueTags {
				if tag.TagId == dataValueTag.TagId {
					return false
				}
			}
		}
	}

	// If the filter value is a slice of interfaces, iterates over the filter value tag ids and the data value tags
	//   and returns false if any of the filter tag ids match any of the data value tags.
	filterValueTagIdsO, tagIdsOk := filterValue.([]interface{})
	if tagIdsOk {
	filterLoop:
		for _, tagIdO := range filterValueTagIdsO {
			for _, dataValueTag := range dataValueTags {
				tagId, err := getFloat(tagIdO)
				if err != nil {
					log.Error().Msgf("could not run convert interface into tagId in filter function (FilterCategoryTypeTag)")
					tagIdsOk = false
					break filterLoop
				}
				if int(tagId) == dataValueTag.TagId {
					return false
				}
			}
		}
	}

	// Converts the tag id to a float and returns false if it fails.
	if !tagsOk && !tagResponsesOk && !tagIdsOk {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeTag) so defaulting to false!")
		return false
	}

	return true
}

func filterCategoryTypeBooleanEq(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if filterValue == nil && dataMap[dataKey] == nil {
		return true
	}
	dataValueBoolean, ok := dataMap[dataKey].(bool)
	if !ok {
		dataValueBooleanPointer, ok := dataMap[dataKey].(*bool)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeBoolean: dataValueBoolean) so defaulting to false instead of a panic!")
			return false
		}
		dataValueBoolean = *dataValueBooleanPointer
	}
	filterValueBoolean, ok := filterValue.(bool)
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: filterValueBoolean) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueBoolean == filterValueBoolean
}

func filterCategoryTypeBooleanNeq(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if filterValue == nil && dataMap[dataKey] == nil {
		return false
	}
	if filterValue != nil && dataMap[dataKey] == nil {
		return true
	}
	if filterValue == nil && dataMap[dataKey] != nil {
		return true
	}
	dataValueBoolean, ok := dataMap[dataKey].(bool)
	if !ok {
		dataValueBooleanPointer, ok := dataMap[dataKey].(*bool)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeBoolean: dataValueBoolean) so defaulting to false instead of a panic!")
			return false
		}
		dataValueBoolean = *dataValueBooleanPointer
	}
	filterValueBoolean, ok := filterValue.(bool)
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: filterValueBoolean) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueBoolean != filterValueBoolean
}
