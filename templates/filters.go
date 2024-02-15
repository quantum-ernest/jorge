package templates

import (
	"fmt"
	"reflect"

	"encoding/xml"
	"github.com/osteele/liquid"
	"github.com/osteele/liquid/evaluator"
	"github.com/osteele/liquid/expressions"
	"time"
)

// a lot of the filters and tags available at jekyll aren't default liquid manually adding them here
// copied from https://github.com/osteele/gojekyll/blob/f1794a874890bfb601cae767a0cce15d672e9058/filters/filters.go
// MIT License: https://github.com/osteele/gojekyll/blob/f1794a874890bfb601cae767a0cce15d672e9058/LICENSE
func loadJekyllFilters(e *liquid.Engine) {
	e.RegisterFilter("filter", filter)
	e.RegisterFilter("group_by", groupByFilter)
	e.RegisterFilter("group_by_exp", groupByExpFilter)
	e.RegisterFilter("sort", sortFilter)
	e.RegisterFilter("where", whereFilter)
	e.RegisterFilter("where_exp", whereExpFilter)
	e.RegisterFilter("xml_escape", xml.Marshal)

	e.RegisterFilter("absolute_url", func(s string) string {
		// FIXME implement after adding a config struct, using the url
		// return utils.URLJoin(c.AbsoluteURL, c.BaseURL, s)
		return s
	})

	e.RegisterFilter("date_to_rfc822", func(date time.Time) string {
		return date.Format(time.RFC822)
		// Out: Mon, 07 Nov 2008 13:07:54 -0800
	})
	e.RegisterFilter("date_to_string", func(date time.Time) string {
		return date.Format("02 Jan 2006")
		// Out: 07 Nov 2008
	})
	e.RegisterFilter("date_to_long_string", func(date time.Time) string {
		return date.Format("02 January 2006")
		// Out: 07 November 2008
	})
	e.RegisterFilter("date_to_xmlschema", func(date time.Time) string {
		return date.Format("2006-01-02T15:04:05-07:00")
		// Out: 2008-11-07T13:07:54-08:00
	})
}

func filter(values []map[string]interface{}, key string) []interface{} {
	var result []interface{}
	for _, value := range values {
		if _, ok := value[key]; ok {
			result = append(result, value)
		}
	}
	return result
}

func groupByExpFilter(array []map[string]interface{}, name string, expr expressions.Closure) ([]map[string]interface{}, error) {
	rt := reflect.ValueOf(array)
	if !(rt.Kind() != reflect.Array || rt.Kind() == reflect.Slice) {
		return nil, nil
	}
	groups := map[interface{}][]interface{}{}
	for i := 0; i < rt.Len(); i++ {
		item := rt.Index(i).Interface()
		key, err := expr.Bind(name, item).Evaluate()
		if err != nil {
			return nil, err
		}
		if group, found := groups[key]; found {
			groups[key] = append(group, item)
		} else {
			groups[key] = []interface{}{item}
		}
	}
	var result []map[string]interface{}
	for k, v := range groups {
		result = append(result, map[string]interface{}{"name": k, "items": v})
	}
	return result, nil
}

func groupByFilter(array []map[string]interface{}, property string) []map[string]interface{} {
	rt := reflect.ValueOf(array)
	if !(rt.Kind() != reflect.Array || rt.Kind() == reflect.Slice) {
		return nil
	}
	groups := map[interface{}][]interface{}{}
	for i := 0; i < rt.Len(); i++ {
		irt := rt.Index(i)
		if irt.Kind() == reflect.Map && irt.Type().Key().Kind() == reflect.String {
			krt := irt.MapIndex(reflect.ValueOf(property))
			if krt.IsValid() && krt.CanInterface() {
				key := krt.Interface()
				if group, found := groups[key]; found {
					groups[key] = append(group, irt.Interface())
				} else {
					groups[key] = []interface{}{irt.Interface()}
				}
			}
		}
	}
	var result []map[string]interface{}
	for k, v := range groups {
		result = append(result, map[string]interface{}{"name": k, "items": v})
	}
	return result
}

func sortFilter(array []interface{}, key interface{}, nilFirst func(bool) bool) []interface{} {
	nf := nilFirst(true)
	result := make([]interface{}, len(array))
	copy(result, array)
	if key == nil {
		evaluator.Sort(result)
	} else {
		// TODO error if key is not a string
		evaluator.SortByProperty(result, key.(string), nf)
	}
	return result
}

func whereExpFilter(array []interface{}, name string, expr expressions.Closure) ([]interface{}, error) {
	rt := reflect.ValueOf(array)
	if rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice {
		return nil, nil
	}
	var result []interface{}
	for i := 0; i < rt.Len(); i++ {
		item := rt.Index(i).Interface()
		value, err := expr.Bind(name, item).Evaluate()
		if err != nil {
			return nil, err
		}
		if value != nil && value != false {
			result = append(result, item)
		}
	}
	return result, nil
}

func whereFilter(array []map[string]interface{}, key string, value interface{}) []interface{} {
	rt := reflect.ValueOf(array)
	if rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice {
		return nil
	}
	var result []interface{}
	for i := 0; i < rt.Len(); i++ {
		item := rt.Index(i)
		if item.Kind() == reflect.Map && item.Type().Key().Kind() == reflect.String {
			attr := item.MapIndex(reflect.ValueOf(key))
			if attr.IsValid() && fmt.Sprint(attr) == value {
				result = append(result, item.Interface())
			}
		}
	}
	return result
}