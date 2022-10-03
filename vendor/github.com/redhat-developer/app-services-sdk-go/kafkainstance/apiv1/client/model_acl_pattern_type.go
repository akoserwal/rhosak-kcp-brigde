/*
 * Kafka Instance API
 *
 * API for interacting with Kafka Instance. Includes Produce, Consume and Admin APIs
 *
 * API version: 0.12.0
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package kafkainstanceclient

import (
	"encoding/json"
	"fmt"
)

// AclPatternType the model 'AclPatternType'
type AclPatternType string

// List of AclPatternType
const (
	ACLPATTERNTYPE_LITERAL AclPatternType = "LITERAL"
	ACLPATTERNTYPE_PREFIXED AclPatternType = "PREFIXED"
)

var allowedAclPatternTypeEnumValues = []AclPatternType{
	"LITERAL",
	"PREFIXED",
}

func (v *AclPatternType) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := AclPatternType(value)
	for _, existing := range allowedAclPatternTypeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid AclPatternType", value)
}

// NewAclPatternTypeFromValue returns a pointer to a valid AclPatternType
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewAclPatternTypeFromValue(v string) (*AclPatternType, error) {
	ev := AclPatternType(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for AclPatternType: valid values are %v", v, allowedAclPatternTypeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v AclPatternType) IsValid() bool {
	for _, existing := range allowedAclPatternTypeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to AclPatternType value
func (v AclPatternType) Ptr() *AclPatternType {
	return &v
}

type NullableAclPatternType struct {
	value *AclPatternType
	isSet bool
}

func (v NullableAclPatternType) Get() *AclPatternType {
	return v.value
}

func (v *NullableAclPatternType) Set(val *AclPatternType) {
	v.value = val
	v.isSet = true
}

func (v NullableAclPatternType) IsSet() bool {
	return v.isSet
}

func (v *NullableAclPatternType) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAclPatternType(val *AclPatternType) *NullableAclPatternType {
	return &NullableAclPatternType{value: val, isSet: true}
}

func (v NullableAclPatternType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAclPatternType) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

