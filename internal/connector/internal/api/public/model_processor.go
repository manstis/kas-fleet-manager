/*
 * Connector Management API
 *
 * Connector Management API is a REST API to manage connectors.
 *
 * API version: 0.1.0
 * Contact: rhosak-support@redhat.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package public

import (
	"time"
)

// Processor struct for Processor
type Processor struct {
	Id              string                `json:"id,omitempty"`
	Kind            string                `json:"kind,omitempty"`
	Href            string                `json:"href,omitempty"`
	Owner           string                `json:"owner,omitempty"`
	CreatedAt       time.Time             `json:"created_at,omitempty"`
	ModifiedAt      time.Time             `json:"modified_at,omitempty"`
	Name            string                `json:"name"`
	NamespaceId     string                `json:"namespace_id"`
	ProcessorTypeId string                `json:"processor_type_id"`
	Channel         Channel               `json:"channel,omitempty"`
	DesiredState    ProcessorDesiredState `json:"desired_state"`
	// Name-value string annotations for resource
	Annotations     map[string]string       `json:"annotations,omitempty"`
	ResourceVersion int64                   `json:"resource_version,omitempty"`
	Kafka           KafkaConnectionSettings `json:"kafka"`
	ServiceAccount  ServiceAccount          `json:"service_account"`
	Definition      map[string]interface{}  `json:"definition,omitempty"`
	ErrorHandler    ErrorHandler            `json:"error_handler,omitempty"`
	Status          ProcessorStatusStatus   `json:"status,omitempty"`
}