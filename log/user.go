package log

import (
	"context"
	"strings"
)

// Key for the map of all user properties added to a context
const UserPropertiesKey = "user"

type UserProperty string

// The following keys are used to add user properties (as a map of UserProperty -> string) to a context for use by loggers
const (
	// Key for user id added to the context in the user properties map
	UserPropertyId UserProperty = "userId"
	// Key for user email added to the context in the user properties map
	UserPropertyEmail UserProperty = "email"
	// Key for user name added to the context in the user properties map
	UserPropertyName UserProperty = "userName"
)

var UserProperties = [...]UserProperty{UserPropertyId, UserPropertyEmail, UserPropertyName}

// ContainsUserProperty checks if a UserProperty value exists in an array of UserProperty.
//
// It takes in two parameters: an array of UserProperty and a UserProperty value to search for.
// It returns a boolean indicating whether or not the value was found in the array.
func ContainsUserProperty(array []UserProperty, value UserProperty) bool {
	for _, a := range array {
		if a == value {
			return true
		}
	}

	return false
}

// GetUserPropertiesMap returns the user properties map from the given context, or nil if no user properties are found.
func GetUserPropertiesMap(ctx context.Context) *map[UserProperty]string {
	if ctx != nil {
		if userProperties := ctx.Value(UserPropertiesKey); userProperties != nil {
			if userPropertiesMap, ok := userProperties.(*map[UserProperty]string); ok {
				return userPropertiesMap
			}
		}
	}

	return nil
}

// GetUserPropertiesString returns a string that contains comma-separated user properties
// that are specified in the given context and userPropertiesToLog slice.
//
// ctx is the context that contains user properties information. userPropertiesToLog
// is a pointer to a slice of UserProperty that specifies which user properties to log.
// A pointer to a string is returned that contains comma-separated user properties.
// If no user properties are specified or the context is nil, nil is returned.
func GetUserPropertiesString(ctx context.Context, userPropertiesToLog *[]UserProperty) *string {
	haveUserToLog := false

	if ctx != nil && userPropertiesToLog != nil && len(*userPropertiesToLog) > 0 {
		if userProperties := ctx.Value(UserPropertiesKey); userProperties != nil {
			if userPropertiesMap, ok := userProperties.(*map[UserProperty]string); ok {
				tempResult := []string{}

				for _, userProperty := range *userPropertiesToLog {
					if userPropertyValue, ok := (*userPropertiesMap)[userProperty]; ok {
						tempResult = append(tempResult, userPropertyValue)
						haveUserToLog = true
					}
				}

				if haveUserToLog {
					result := strings.Join(tempResult, ", ")
					return &result
				}
			}
		}
	}

	return nil
}
