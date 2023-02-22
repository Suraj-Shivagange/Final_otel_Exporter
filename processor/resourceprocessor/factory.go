package resourceprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/vunetsystems/opentelemetry-collector/processor/resourceprocessor/attraction"
)

const (
	// The value of "type" key in configuration.
	typeStr = "resource"
	// The stability level of the processor.
	stability = component.StabilityLevelBeta
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Resource processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, stability),
		processor.WithMetrics(createMetricsProcessor, stability),
		processor.WithLogs(createLogsProcessor, stability))
}

// Note: This isn't a valid configuration because the processor would do no work.
func createDefaultConfig() component.Config {
	return &Config{}
}

func createTracesProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces) (processor.Traces, error) {
	attrProc, err := attraction.NewAttrProc(&attraction.Settings{Actions: cfg.(*Config).AttributesActions})
	if err != nil {
		return nil, err
	}
	proc := &resourceProcessor{logger: set.Logger, attrProc: attrProc}
	return processorhelper.NewTracesProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics) (processor.Metrics, error) {
	attrProc, err := attraction.NewAttrProc(&attraction.Settings{Actions: cfg.(*Config).AttributesActions})
	if err != nil {
		return nil, err
	}
	proc := &resourceProcessor{logger: set.Logger, attrProc: attrProc}
	return processorhelper.NewMetricsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createLogsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Logs) (processor.Logs, error) {
	attrProc, err := attraction.NewAttrProc(&attraction.Settings{Actions: cfg.(*Config).AttributesActions})
	if err != nil {
		return nil, err
	}
	proc := &resourceProcessor{logger: set.Logger, attrProc: attrProc}
	return processorhelper.NewLogsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}
