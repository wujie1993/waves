// Code generated by go-swagger; //CARE! edited: OpDesc

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ResponseMessage response message
// swagger:model ResponseMessage
type ResponseMessage struct {

	// data
	Data interface{} `json:"Data,omitempty"`

	// op code
	OpCode string `json:"OpCode,omitempty"`

	// op message
	// OpMessage string `json:"OpMessage,omitempty"`

	OpDesc string `json:"OpDesc,omitempty"`
}

// Validate validates this response message
func (m *ResponseMessage) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ResponseMessage) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ResponseMessage) UnmarshalBinary(b []byte) error {
	var res ResponseMessage
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
