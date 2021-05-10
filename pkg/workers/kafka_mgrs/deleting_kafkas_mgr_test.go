package kafka_mgrs

import (
	"testing"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/config"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/api"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/errors"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/services"
)

func TestDeletingKafkaManager(t *testing.T) {
	type fields struct {
		kafkaService  services.KafkaService
		quotaService  services.QuotaService
		configService services.ConfigService
	}
	type args struct {
		kafka *api.KafkaRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful reconcile",
			args: args{
				kafka: &api.KafkaRequest{},
			},
			fields: fields{
				kafkaService: &services.KafkaServiceMock{
					DeleteFunc: func(kafkaRequest *api.KafkaRequest) *errors.ServiceError {
						//return errors.GeneralError("test")
						return nil
					},
				},
				quotaService: &services.QuotaServiceMock{
					DeleteQuotaFunc: func(id string) *errors.ServiceError {
						return nil
					},
				},
				configService: services.NewConfigService(config.ApplicationConfig{
					Kafka: config.NewKafkaConfig(),
				}),
			},
		},
		{
			name: "failed reconcile",
			args: args{
				kafka: &api.KafkaRequest{},
			},
			fields: fields{
				kafkaService: &services.KafkaServiceMock{
					DeleteFunc: func(kafkaRequest *api.KafkaRequest) *errors.ServiceError {
						return errors.GeneralError("test")
					},
				},
				quotaService: &services.QuotaServiceMock{
					DeleteQuotaFunc: func(id string) *errors.ServiceError {
						return nil
					},
				},
				configService: services.NewConfigService(config.ApplicationConfig{
					Kafka: config.NewKafkaConfig(),
				}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &DeletingKafkaManager{
				kafkaService:  tt.fields.kafkaService,
				quotaService:  tt.fields.quotaService,
				configService: tt.fields.configService,
			}
			if err := k.reconcileDeprovisioningRequest(tt.args.kafka); (err != nil) != tt.wantErr {
				t.Errorf("reconcileDeprovisioningRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
