package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Adapptor/service/log"
)

type Map map[string]interface{}

func (m *Map) GetString(key string) string {
	var vstr string

	if value, ok := (*m)[key]; ok {
		vstr, _ = value.(string)
	}

	return vstr
}

func ReadJsonMap(reader io.Reader) (Map, error) {
	decoder := json.NewDecoder(reader)
	var body Map
	err := decoder.Decode(&body)
	return body, err
}

type MapError struct {
	s string
}

var Nil = MapError{s: "service.MapNil"}
var Remove = MapError{s: "service.MapRemove"}

// UpdatePath updates the value at the given path
func (m *Map) UpdatePath(path string, updateValue interface{}, ctx context.Context) (interface{}, error) {
	return m.traversePath(path, updateValue, true, ctx)
}

// UpdateExistingPath updates an existing value at the given path
func (m *Map) UpdateExistingPath(path string, updateValue interface{}, ctx context.Context) (interface{}, error) {
	return m.traversePath(path, updateValue, false, ctx)
}

// QueryPath queries the value at the given path
func (m *Map) QueryPath(path string, ctx context.Context) (interface{}, error) {
	return m.traversePath(path, nil, false, ctx)
}

// Traverse the given slash separated path and update if a value is provided,
// returning the updated value.  Otherwise return the value at the given path.
//
// updateValue can be a valid JSON value (map, string, number, bool). An
// updateValue of nil requests a query, Nil requests a value of nil, Remove
// requests a deletion
//
// If createPaths is true, any missing path components are initialized as
// empty maps
func (m *Map) traversePath(path string, updateValue interface{}, createPaths bool, ctx context.Context) (interface{}, error) {
	path = strings.TrimPrefix(path, "/")
	components := strings.Split(path, "/")

	log.Log(log.Debug, fmt.Sprintf("Traversing path %v with value %v", path, updateValue), nil, ctx)

	if len(components) == 1 && components[0] == "" && updateValue != nil {
		//  Complete replacement of root map, updateValue must be a generic map
		*m = Map(updateValue.(map[string]interface{}))
		return m, nil
	}

	var lastIndex = len(components) - 1

	ref := *m
	var child interface{} = nil

	for i, component := range components {
		var ok bool

		if component == "" {
			return nil, fmt.Errorf("empty component encountered in path %v", path)
		}

		isUpdate := updateValue != nil

		if i == lastIndex && isUpdate {
			log.Log(log.Debug, fmt.Sprintf("Updating component %v value %+v", component, updateValue), nil, ctx)

			var jsonUpdateValue = updateValue
			if updateValue == Nil {
				jsonUpdateValue = nil
			} else if updateValue == Remove {
				delete(ref, component)
				return Remove, nil
			}

			ref[component] = jsonUpdateValue
			return ref[component], nil
		} else {
			child, ok = ref[component]
			//  Error if this child is not found
			if !ok {
				if createPaths && isUpdate {
					log.Log(log.Debug, fmt.Sprintf("Creating path for component %v", component), nil, ctx)
					newPath := map[string]interface{}{}
					ref[component] = newPath
					ref = newPath
					continue
				} else {
					return nil, fmt.Errorf("child component %v of path %v not found", component, path)
				}
			}

			if i == lastIndex && !isUpdate {
				//  Return the queried value
				log.Log(log.Debug, fmt.Sprintf("Returning query value %+v", child), nil, ctx)
				return child, nil
			}

			//  Keep going - child must be a map to enable further traversal
			ref, ok = child.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("child component %v of path %v is not a map", component, path)
			}
		}
	}

	//  XXX Shouldn't get here
	return nil, fmt.Errorf("unexpected return from TraversePath %v", path)
}
