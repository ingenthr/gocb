package gocbcore

import (
	"encoding/json"
	"expvar"
	"fmt"
)

type PipelineState int

const (
	PipelineStateActive = PipelineState(1)
	PipelineStateClosed = PipelineState(1)
)

var (
	goCbMetrics *expvar.Map
)

func init() {
	goCbMetrics = expvar.NewMap("gocb")
}

type PipelineMetrics struct {
	Name           string
	QueueSize      int
	State          PipelineState
	PacketsWritten int64
	PacketsRead    int64
}

type AgentMetrics struct {
	WaitQueueSize int
	DeadQueueSize int

	ActiveServers  []*PipelineMetrics
	PendingServers []*PipelineMetrics
}

func (a AgentMetrics) String() string {
	bytes, err := json.Marshal(a)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(bytes)
}

func pipelineMetrics(pipeline *memdPipeline) *PipelineMetrics {
	metrics := &PipelineMetrics{}

	metrics.Name = pipeline.address
	metrics.QueueSize = len(pipeline.queue.reqsCh)
	metrics.PacketsWritten = pipeline.packetsWritten
	metrics.PacketsRead = pipeline.packetsRead

	if pipeline.isClosed {
		metrics.State = PipelineStateClosed
	} else {
		metrics.State = PipelineStateActive
	}

	return metrics
}

// Experimental
// Retrieves various internal metrics related to the routing
//   and dispatching of operations.
func (agent *Agent) PollMetrics() *AgentMetrics {
	metrics := &AgentMetrics{}

	routingInfo := agent.routingInfo.get()
	if routingInfo != nil {
		if routingInfo.waitQueue != nil {
			metrics.WaitQueueSize = len(routingInfo.waitQueue.reqsCh)
		} else {
			metrics.WaitQueueSize = -1
		}
		if routingInfo.deadQueue != nil {
			metrics.DeadQueueSize = len(routingInfo.deadQueue.reqsCh)
		} else {
			metrics.DeadQueueSize = -1
		}

		for _, pipeline := range routingInfo.servers {
			if pipeline != nil {
				metrics.ActiveServers = append(metrics.ActiveServers, pipelineMetrics(pipeline))
			}
		}

		for _, pipeline := range routingInfo.pendingServers {
			if pipeline != nil {
				metrics.PendingServers = append(metrics.PendingServers, pipelineMetrics(pipeline))
			}
		}
	}

	expvarKeyName := fmt.Sprintf("AgentMetrics-%p", agent)
	goCbMetrics.Set(expvarKeyName, metrics)

	return metrics
}
