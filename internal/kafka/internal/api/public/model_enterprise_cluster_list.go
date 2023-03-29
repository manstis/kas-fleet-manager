/*
 * Kafka Management API
 *
 * Kafka Management API is a REST API to manage Kafka instances
 *
 * API version: 1.16.0
 * Contact: rhosak-support@redhat.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package public

// EnterpriseClusterList struct for EnterpriseClusterList
type EnterpriseClusterList struct {
	Kind  string              `json:"kind"`
	Page  int32               `json:"page"`
	Size  int32               `json:"size"`
	Total int32               `json:"total"`
	Items []EnterpriseCluster `json:"items"`
}
