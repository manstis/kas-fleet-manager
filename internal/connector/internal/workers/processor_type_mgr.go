package workers

import (
	"encoding/json"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/api/dbapi"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/config"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/services"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/server"
	"github.com/pkg/errors"
	"reflect"
	"sync"
	"time"

	serviceError "github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/errors"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/workers"

	"github.com/golang/glog"
	"github.com/google/uuid"
)

// ProcessorTypeManager represents a connector manager that reconciles connector types at startup
type ProcessorTypeManager struct {
	workers.BaseWorker
	processorTypesService services.ProcessorTypesService
	startupReconcileDone  bool
	startupReconcileWG    sync.WaitGroup
}

// NewApiServerReadyConditionForProcessors is used to inject a server.ApiServerReadyCondition into the server.ApiServer
// so that it waits for the ProcessorTypeManager to have completed a startup reconcile before accepting http requests.
func NewApiServerReadyConditionForProcessors(pm *ProcessorTypeManager) server.ApiServerReadyCondition {
	return &pm.startupReconcileWG
}

// NewProcessorTypeManager creates a new processor type manager
func NewProcessorTypeManager(
	processorTypesService services.ProcessorTypesService,
	reconciler workers.Reconciler,
) *ProcessorTypeManager {
	result := &ProcessorTypeManager{
		BaseWorker: workers.BaseWorker{
			Id:         uuid.New().String(),
			WorkerType: "processor_type",
			Reconciler: reconciler,
		},
		processorTypesService: processorTypesService,
		startupReconcileDone:  false,
	}

	// The release of this waiting group signal the http service to start serving request
	// this needs to be done across multiple instances of fleetmanager running,
	// and yet just one instance of ProcessorTypeManager of those multiple fleet manager will run the reconcile loop.
	// The release of the waiting group must then be done outside the reconcile loop,
	// the condition is checked in runStartupReconcileCheckWorker().
	result.startupReconcileWG.Add(1)

	// Mark startupReconcileWG as done in a separate goroutine instead of in worker reconcile
	// this is required to allow multiple instances of fleetmanager to startup.
	result.runStartupReconcileCheckWorker()
	return result
}

// Start initializes the connector manager to reconcile connector requests
func (k *ProcessorTypeManager) Start() {
	k.StartWorker(k)
}

// Stop causes the process for reconciling connector requests to stop.
func (k *ProcessorTypeManager) Stop() {
	k.StopWorker(k)
}

// HasTerminated indicates whether the worker should be stopped and terminated
func (k *ProcessorTypeManager) HasTerminated() bool {
	return k.startupReconcileDone
}

func (k *ProcessorTypeManager) Reconcile() []error {
	if !k.startupReconcileDone {
		glog.V(5).Infoln("Reconciling startup processor catalog updates...")

		// the assumption here is that this runs on one instance only of fleetmanager,
		// runs only at startup and while requests are not being served
		// this call handles types that are not in catalog anymore,
		// removing unused types and marking used types as deprecated
		if err := k.processorTypesService.DeleteOrDeprecateRemovedTypes(); err != nil {
			return []error{err}
		}

		// We only need to reconcile channel updates once per process startup since,
		// configured channel settings are only loaded on startup.
		// These operations, once completed successfully, make the condition at runStartupReconcileCheckWorker() to pass
		// practically starting the serving of requests from the service.
		// IMPORTANT: Everything that should run before the first request is served should happen before this
		if err := k.processorTypesService.ForEachProcessorCatalogEntry(k.ReconcileProcessorCatalog); err != nil {
			return []error{err}
		}

		if err := k.processorTypesService.CleanupDeployments(); err != nil {
			return []error{err}
		}

		k.startupReconcileDone = true
		glog.V(5).Infoln("Catalog updates processed")
	}

	return nil
}

func (k *ProcessorTypeManager) ReconcileProcessorCatalog(processorTypeId string, channel string, processorChannelConfig *config.ProcessorChannelConfig) *serviceError.ServiceError {
	processorShardMetadata := dbapi.ProcessorShardMetadata{
		ProcessorTypeId: processorTypeId,
		Channel:         channel,
	}

	var err error
	processorShardMetadata.Revision, err = GetProcessorShardMetadataRevision(processorChannelConfig.ShardMetadata)
	if err != nil {
		return serviceError.GeneralError("Failed to convert Processor Type %s, Channel %s. Error in loaded Processor Type Shard Metadata %+v: %v", processorTypeId, channel, processorChannelConfig.ShardMetadata, err.Error())
	}
	processorShardMetadata.ShardMetadata, err = json.Marshal(processorChannelConfig.ShardMetadata)
	if err != nil {
		return serviceError.GeneralError("Failed to convert Processor Type %s, Channel %s: %v", processorTypeId, channel, err.Error())
	}

	// We store processor type channels so that we can track changes and trigger redeployment of
	// associated connectors upon connector type channel changes.
	_, serr := k.processorTypesService.PutProcessorShardMetadata(&processorShardMetadata)
	if serr != nil {
		return serr
	}

	return nil
}

func GetProcessorShardMetadataRevision(processorShardMetadata map[string]interface{}) (int64, error) {
	revision, processorRevisionFound := processorShardMetadata["processor_revision"]
	if processorRevisionFound {
		floatRevision, isfloat64 := revision.(float64)
		if isfloat64 {
			return int64(floatRevision), nil
		} else {
			return 0, errors.Errorf("processor_revision in Shard Metadata was not an int but a %v", reflect.TypeOf(revision).Kind())
		}
	} else {
		return 0, errors.Errorf("processor_revision not found in Shard Metadata")
	}
}

func (k *ProcessorTypeManager) runStartupReconcileCheckWorker() {
	go func() {
		for !k.startupReconcileDone {
			glog.V(5).Infoln("Waiting for startup processor catalog updates...")
			// Check that ProcessorTypes in the current configured catalog have the same checksums as the ones stored in the db (comparing them by id).
			done, err := k.processorTypesService.CatalogEntriesReconciled()
			if err != nil {
				glog.Errorf("Error checking catalog entry checksums: %s", err)
			} else if done {
				k.startupReconcileDone = true
			} else {
				// wait another 5 seconds to check
				time.Sleep(checkCatalogEntriesDuration)
			}
		}
		glog.V(5).Infoln("Wait for processor catalog updates done!")
		k.startupReconcileWG.Done()
	}()
}