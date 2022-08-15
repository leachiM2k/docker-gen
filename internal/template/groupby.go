package template

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nginx-proxy/docker-gen/internal/context"
)

// Generalized groupBy function
func generalizedGroupBy(funcName string, entries interface{}, getValue func(interface{}) (interface{}, error), addEntry func(map[string][]interface{}, interface{}, interface{})) (map[string][]interface{}, error) {
	entriesVal, err := getArrayValues(funcName, entries)

	if err != nil {
		return nil, err
	}

	groups := make(map[string][]interface{})
	for i := 0; i < entriesVal.Len(); i++ {
		v := reflect.Indirect(entriesVal.Index(i)).Interface()
		value, err := getValue(v)
		if err != nil {
			return nil, err
		}
		if value != nil {
			addEntry(groups, value, v)
		}
	}
	return groups, nil
}

func generalizedGroupByKey(funcName string, entries interface{}, key string, addEntry func(map[string][]interface{}, interface{}, interface{})) (map[string][]interface{}, error) {
	getKey := func(v interface{}) (interface{}, error) {
		return deepGet(v, key), nil
	}
	return generalizedGroupBy(funcName, entries, getKey, addEntry)
}

func groupByMulti(entries interface{}, key, sep string) (map[string][]interface{}, error) {
	return generalizedGroupByKey("groupByMulti", entries, key, func(groups map[string][]interface{}, value interface{}, v interface{}) {
		items := strings.Split(value.(string), sep)
		for _, item := range items {
			groups[item] = append(groups[item], v)
		}
	})
}

// groupBy groups a generic array or slice by the path property key
func groupBy(entries interface{}, key string) (map[string][]interface{}, error) {
	return generalizedGroupByKey("groupBy", entries, key, func(groups map[string][]interface{}, value interface{}, v interface{}) {
		groups[value.(string)] = append(groups[value.(string)], v)
	})
}

// groupByKeys is the same as groupBy but only returns a list of keys
func groupByKeys(entries interface{}, key string) ([]string, error) {
	keys, err := generalizedGroupByKey("groupByKeys", entries, key, func(groups map[string][]interface{}, value interface{}, v interface{}) {
		groups[value.(string)] = append(groups[value.(string)], v)
	})

	if err != nil {
		return nil, err
	}

	ret := []string{}
	for k := range keys {
		ret = append(ret, k)
	}
	return ret, nil
}

// groupByLabel is the same as groupBy but over a given label
func groupByLabel(entries interface{}, label string) (map[string][]interface{}, error) {
	getLabel := func(v interface{}) (interface{}, error) {
		if container, ok := v.(context.RuntimeContainer); ok {
			if value, ok := container.Labels[label]; ok {
				return value, nil
			}
			return nil, nil
		}
		return nil, fmt.Errorf("must pass an array or slice of RuntimeContainer to 'groupByLabel'; received %v", v)
	}
	return generalizedGroupBy("groupByLabel", entries, getLabel, func(groups map[string][]interface{}, value interface{}, v interface{}) {
		groups[value.(string)] = append(groups[value.(string)], v)
	})
}

// splitKeyValuePairs splits a input string into a map of key value pairs, first string is split by listSep into list items, then each list item is split by kvpSep into key value pair
// if a list item does not contai the kvpSep a defaultKey can be provided, where these values are grouped, or if omitted these values are used as key and value
func splitKeyValuePairs(input string, listSep string, kvpSep string, defaultKey ...string) map[string]string {
	keyValuePairs := strings.Split(input, listSep)

	output := map[string]string{}
	for _, kvp := range keyValuePairs {
		var key string
		var value string
		if strings.Contains(kvp, kvpSep) {
			splitted := strings.Split(kvp, kvpSep)
			key = splitted[0]
			value = splitted[1]
		} else if len(defaultKey) == 0 || defaultKey[0] == "" {
			// no key found, no default key specified
			key = kvp
			value = kvp
		} else {
			// no key found, use default key specified instead
			key = defaultKey[0]
			value = kvp
		}

		output[key] = value
	}

	return output
}

// groupByMultiKeyValuePairs similar to groupByMulti, but the key value ist split into a list (delimited by listSep) of key value pairs (seperated by kvpSep: <key>kvpSep<value, e.g key1=value1>)
// An array or slice entry will show up in the output map under all of the list key value pair keys
func groupByMultiKeyValuePairs(entries interface{}, key, listSep string, kvpSep string, defaultKey string) (map[string][]interface{}, error) {
	return generalizedGroupByKey("groupByMultiKeyValuePairs", entries, key, func(groups map[string][]interface{}, value interface{}, v interface{}) {

		keyValuePairs := splitKeyValuePairs(value.(string), listSep, kvpSep, defaultKey)
		for key := range keyValuePairs {
			groups[key] = append(groups[key], v)
		}
	})
}

