/*
 * Kafka Service Fleet Manager
 *
 * Kafka Service Fleet Manager is a Rest API to manage Kafka instances.
 *
 * API version: 1.8.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package public

// Error struct for Error
type Error struct {
	Id          string `json:"id,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Href        string `json:"href,omitempty"`
	Code        string `json:"code,omitempty"`
	Reason      string `json:"reason,omitempty"`
	OperationId string `json:"operation_id,omitempty"`
}
