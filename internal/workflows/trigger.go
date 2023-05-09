package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lightning"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/internal/workflow_helpers"
)

type workflowVersionNodeIdType int
type stageType int

func ProcessWorkflow(ctx context.Context, db *sqlx.DB,
	workflowTriggerNode WorkflowNode,
	reference string,
	events []any) error {

	workflowNodeInputCache := make(map[workflowVersionNodeIdType]map[workflow_helpers.WorkflowParameterLabel]string)
	workflowNodeInputByReferenceIdCache := make(map[workflowVersionNodeIdType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string)
	workflowNodeOutputCache := make(map[workflowVersionNodeIdType]map[workflow_helpers.WorkflowParameterLabel]string)
	workflowNodeOutputByReferenceIdCache := make(map[workflowVersionNodeIdType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string)
	workflowStageOutputCache := make(map[stageType]map[workflow_helpers.WorkflowParameterLabel]string)
	workflowStageOutputByReferenceIdCache := make(map[stageType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string)

	select {
	case <-ctx.Done():
		return errors.New(fmt.Sprintf("Context terminated for WorkflowVersionId: %v", workflowTriggerNode.WorkflowVersionId))
	default:
	}

	if workflowTriggerNode.Status != WorkflowNodeActive {
		return nil
	}

	workflowNodeStatus := make(map[int]core.Status)
	workflowNodeStatus[workflowTriggerNode.WorkflowVersionNodeId] = core.Active

	var eventChannelIds []int
	for _, event := range events {
		channelBalanceEvent, ok := event.(core.ChannelBalanceEvent)
		if ok {
			eventChannelIds = append(eventChannelIds, channelBalanceEvent.ChannelId)
		}
		channelEvent, ok := event.(core.ChannelEvent)
		if ok {
			eventChannelIds = append(eventChannelIds, channelEvent.ChannelId)
		}
	}
	marshalledEventChannelIdsFromEvents, err := json.Marshal(eventChannelIds)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal eventChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}
	marshalledEvents, err := json.Marshal(events)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal events for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}

	var allChannelIds []int
	torqNodeIds := cache.GetAllTorqNodeIds()
	for _, torqNodeId := range torqNodeIds {
		// Force Response because we don't care about balance accuracy
		channelIdsByNode := cache.GetChannelStateNotSharedChannelIds(torqNodeId, true)
		allChannelIds = append(allChannelIds, channelIdsByNode...)
	}
	marshalledAllChannelIds, err := json.Marshal(allChannelIds)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal allChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}

	workflowVersionNodes, err := GetWorkflowVersionNodesByStage(db, workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.Stage)
	if err != nil {
		return errors.Wrapf(err, "Failed to obtain workflow nodes for WorkflowVersionId: %v (stage: %v)",
			workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.Stage)
	}
	initializeInputCache(workflowVersionNodes, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
		workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
		allChannelIds, marshalledEventChannelIdsFromEvents, marshalledAllChannelIds, marshalledEvents,
		workflowTriggerNode, workflowStageOutputCache, workflowStageOutputByReferenceIdCache)

	switch workflowTriggerNode.Type {
	case workflow_helpers.WorkflowNodeIntervalTrigger:
		log.Debug().Msgf("Interval Trigger Fired for WorkflowVersionNodeId: %v",
			workflowTriggerNode.WorkflowVersionNodeId)
	case workflow_helpers.WorkflowNodeCronTrigger:
		log.Debug().Msgf("Cron Trigger Fired for WorkflowVersionNodeId: %v",
			workflowTriggerNode.WorkflowVersionNodeId)
	case workflow_helpers.WorkflowNodeChannelBalanceEventTrigger:
		log.Debug().Msgf("Channel Balance Event Trigger Fired for WorkflowVersionNodeId: %v",
			workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowVersionNodeIdType(workflowTriggerNode.WorkflowVersionNodeId)][workflow_helpers.WorkflowParameterLabelChannels] = string(marshalledEventChannelIdsFromEvents)
	case workflow_helpers.WorkflowNodeChannelOpenEventTrigger:
		log.Debug().Msgf("Channel Open Event Trigger Fired for WorkflowVersionNodeId: %v",
			workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowVersionNodeIdType(workflowTriggerNode.WorkflowVersionNodeId)][workflow_helpers.WorkflowParameterLabelChannels] = string(marshalledEventChannelIdsFromEvents)
	case workflow_helpers.WorkflowNodeChannelCloseEventTrigger:
		log.Debug().Msgf("Channel Close Event Trigger Fired for WorkflowVersionNodeId: %v",
			workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowVersionNodeIdType(workflowTriggerNode.WorkflowVersionNodeId)][workflow_helpers.WorkflowParameterLabelChannels] = string(marshalledEventChannelIdsFromEvents)
	case workflow_helpers.WorkflowTrigger:
		log.Debug().Msgf("Trigger Fired for WorkflowVersionNodeId: %v",
			workflowTriggerNode.WorkflowVersionNodeId)
	case workflow_helpers.WorkflowNodeManualTrigger:
		log.Debug().Msgf("Manual Trigger Fired for WorkflowVersionNodeId: %v",
			workflowTriggerNode.WorkflowVersionNodeId)
	}

	done := false
	iteration := 0
	var processStatus core.Status
	for !done {
		iteration++
		if iteration > 100 {
			return errors.New(fmt.Sprintf("Infinite loop for WorkflowVersionId: %v", workflowTriggerNode.WorkflowVersionId))
		}
		done = true
		for _, workflowVersionNode := range workflowVersionNodes {
			processStatus, err = processWorkflowNode(ctx, db, workflowVersionNode, workflowVersionNodes, workflowTriggerNode,
				workflowNodeStatus, reference, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
				workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
				workflowStageOutputCache, workflowStageOutputByReferenceIdCache)
			if err != nil {
				return errors.Wrapf(err, "Failed to process workflow nodes for WorkflowVersionId: %v (stage: %v)",
					workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.Stage)
			}
			if processStatus == core.Active {
				done = false
			}
		}
	}

	workflowStageTriggerNodes, err := GetActiveSortedStageTriggerNodeForWorkflowVersionId(db,
		workflowTriggerNode.WorkflowVersionId)
	if err != nil {
		return errors.Wrapf(err, "Failed to obtain stage workflow trigger nodes for WorkflowVersionId: %v",
			workflowTriggerNode.WorkflowVersionId)
	}

	for _, workflowStageTriggerNode := range workflowStageTriggerNodes {
		workflowNodeInputCache = make(map[workflowVersionNodeIdType]map[workflow_helpers.WorkflowParameterLabel]string)
		workflowNodeInputByReferenceIdCache = make(map[workflowVersionNodeIdType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string)
		workflowNodeOutputCache = make(map[workflowVersionNodeIdType]map[workflow_helpers.WorkflowParameterLabel]string)
		workflowNodeOutputByReferenceIdCache = make(map[workflowVersionNodeIdType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string)

		workflowVersionNodes, err = GetWorkflowVersionNodesByStage(db, workflowTriggerNode.WorkflowVersionId, workflowStageTriggerNode.Stage)
		if err != nil {
			return errors.Wrapf(err, "Failed to obtain workflow nodes for WorkflowVersionId: %v (stage: %v)",
				workflowTriggerNode.WorkflowVersionId, workflowStageTriggerNode.Stage)
		}

		initializeInputCache(workflowVersionNodes, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
			workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
			allChannelIds, marshalledEventChannelIdsFromEvents, marshalledAllChannelIds, marshalledEvents,
			workflowStageTriggerNode, workflowStageOutputCache, workflowStageOutputByReferenceIdCache)
		done = false
		iteration = 0
		for !done {
			iteration++
			if iteration > 100 {
				return errors.New(fmt.Sprintf("Infinite loop for WorkflowVersionId: %v", workflowStageTriggerNode.WorkflowVersionId))
			}
			done = true
			for _, workflowVersionNode := range workflowVersionNodes {
				processStatus, err = processWorkflowNode(ctx, db, workflowVersionNode, workflowVersionNodes, workflowTriggerNode,
					workflowNodeStatus, reference, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
					workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
					workflowStageOutputCache, workflowStageOutputByReferenceIdCache)
				if err != nil {
					return errors.Wrapf(err, "Failed to process workflow nodes for WorkflowVersionId: %v (stage: %v)",
						workflowTriggerNode.WorkflowVersionId, workflowStageTriggerNode.Stage)
				}
				if processStatus == core.Active {
					done = false
				}
			}
		}
	}
	return nil
}

func initializeInputCache(workflowVersionNodes []WorkflowNode,
	workflowNodeInputCache map[workflowVersionNodeIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowNodeInputByReferenceIdCache map[workflowVersionNodeIdType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowNodeOutputCache map[workflowVersionNodeIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowNodeOutputByReferenceIdCache map[workflowVersionNodeIdType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	allChannelIds []int,
	marshalledChannelIdsFromEvents []byte,
	marshalledAllChannelIds []byte,
	marshalledEvents []byte,
	workflowStageTriggerNode WorkflowNode,
	workflowStageOutputCache map[stageType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowStageOutputByReferenceIdCache map[stageType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string) {

	for _, workflowVersionNode := range workflowVersionNodes {
		if workflowNodeInputCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)] == nil {
			workflowNodeInputCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)] = make(map[workflow_helpers.WorkflowParameterLabel]string)
		}
		if workflowNodeOutputCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)] == nil {
			workflowNodeOutputCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)] = make(map[workflow_helpers.WorkflowParameterLabel]string)
		}
		workflowNodeInputCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][workflow_helpers.WorkflowParameterLabelEventChannels] = string(marshalledChannelIdsFromEvents)
		workflowNodeInputCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][workflow_helpers.WorkflowParameterLabelAllChannels] = string(marshalledAllChannelIds)
		workflowNodeInputCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][workflow_helpers.WorkflowParameterLabelEvents] = string(marshalledEvents)

		if workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)] == nil {
			workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)] = make(map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string)
		}
		if workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)] == nil {
			workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)] = make(map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string)
		}
		for _, channelId := range allChannelIds {
			if workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelIdType(channelId)] == nil {
				workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelIdType(channelId)] = make(map[workflow_helpers.WorkflowParameterLabel]string)
			}
			if workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelIdType(channelId)] == nil {
				workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelIdType(channelId)] = make(map[workflow_helpers.WorkflowParameterLabel]string)
			}
			workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelIdType(channelId)][workflow_helpers.WorkflowParameterLabelEventChannels] = string(marshalledChannelIdsFromEvents)
			workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelIdType(channelId)][workflow_helpers.WorkflowParameterLabelAllChannels] = string(marshalledAllChannelIds)
			workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelIdType(channelId)][workflow_helpers.WorkflowParameterLabelEvents] = string(marshalledEvents)
		}
		if workflowStageTriggerNode.Stage > 0 {
			for label, value := range workflowStageOutputCache[stageType(workflowStageTriggerNode.Stage-1)] {
				workflowNodeInputCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][label] = value
			}
			for channelId, labelValueMap := range workflowStageOutputByReferenceIdCache[stageType(workflowStageTriggerNode.Stage-1)] {
				if workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelId] == nil {
					workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelId] = make(map[workflow_helpers.WorkflowParameterLabel]string)
				}
				for label, value := range labelValueMap {
					workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowVersionNode.WorkflowVersionNodeId)][channelId][label] = value
				}
			}
		}
	}
}

// processWorkflowNode
// workflowNodeInputCache: map[workflowVersionNodeId][channelId][label]value
// workflowNodeOutputCache: map[workflowVersionNodeId][channelId][label]value
// workflowStageExitConfigurationCache: map[stage][channelId][label]value
func processWorkflowNode(ctx context.Context, db *sqlx.DB,
	workflowNode WorkflowNode,
	workflowNodes []WorkflowNode,
	workflowTriggerNode WorkflowNode,
	workflowNodeStatus map[int]core.Status,
	reference string,
	workflowNodeInputCache map[workflowVersionNodeIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowNodeInputByReferenceIdCache map[workflowVersionNodeIdType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowNodeOutputCache map[workflowVersionNodeIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowNodeOutputByReferenceIdCache map[workflowVersionNodeIdType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowStageOutputCache map[stageType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowStageOutputByReferenceIdCache map[stageType]map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string) (core.Status, error) {

	select {
	case <-ctx.Done():
		return core.Inactive, errors.New(fmt.Sprintf("Context terminated for WorkflowVersionId: %v", workflowNode.WorkflowVersionId))
	default:
	}

	if workflowNode.Status != WorkflowNodeActive {
		return core.Deleted, nil
	}

	status, statusExists := workflowNodeStatus[workflowNode.WorkflowVersionNodeId]
	if statusExists && status == core.Active {
		// When the node is in the cache and active then it's already been processed successfully
		return core.Deleted, nil
	}

	if workflow_helpers.IsWorkflowNodeTypeGrouped(workflowNode.Type) {
		workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = core.Active
		return core.Deleted, nil
	}

	parentLinkedInputs := make(map[workflow_helpers.WorkflowParameterLabel][]WorkflowNode)
	for parentWorkflowNodeLinkId, parentWorkflowNode := range workflowNode.ParentNodes {
		parentLink := workflowNode.LinkDetails[parentWorkflowNodeLinkId]
		parentLinkedInputs[parentLink.ChildInput] = append(parentLinkedInputs[parentLink.ChildInput], *parentWorkflowNode)
	}
linkedInputLoop:
	for _, parentWorkflowNodesByInput := range parentLinkedInputs {
		for _, parentWorkflowNode := range parentWorkflowNodesByInput {
			status, statusExists = workflowNodeStatus[parentWorkflowNode.WorkflowVersionNodeId]
			if statusExists && status == core.Active {
				continue linkedInputLoop
			}
		}
		// Not all inputs are available yet
		return core.Pending, nil
	}

	inputs := workflowNodeInputCache[workflowVersionNodeIdType(workflowNode.WorkflowVersionNodeId)]
	inputsByReferenceId := workflowNodeInputByReferenceIdCache[workflowVersionNodeIdType(workflowNode.WorkflowVersionNodeId)]
	outputs := workflowNodeOutputCache[workflowVersionNodeIdType(workflowNode.WorkflowVersionNodeId)]
	outputsByReferenceId := workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdType(workflowNode.WorkflowVersionNodeId)]

	for parentWorkflowNodeLinkId, parentWorkflowNode := range workflowNode.ParentNodes {
		parentLink := workflowNode.LinkDetails[parentWorkflowNodeLinkId]
		parentOutputValue, labelExists := workflowNodeOutputCache[workflowVersionNodeIdType(parentWorkflowNode.WorkflowVersionNodeId)][parentLink.ParentOutput]
		if labelExists {
			inputs[parentLink.ChildInput] = parentOutputValue
			outputs[parentLink.ChildInput] = parentOutputValue
		}
		for referencId, labelValueMap := range workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdType(parentWorkflowNode.WorkflowVersionNodeId)] {
			parentOutputValueByReferenceId, labelByReferenceIdExists := labelValueMap[parentLink.ParentOutput]
			if labelByReferenceIdExists {
				inputsByReferenceId[referencId][parentLink.ChildInput] = parentOutputValueByReferenceId
				outputsByReferenceId[referencId][parentLink.ChildInput] = parentOutputValueByReferenceId
			}
			for _, workflowNodeParameterLabelEnforced := range workflow_helpers.GetWorkflowParameterLabelsEnforced() {
				parentByReferenceId, parentByReferenceIdExists := labelValueMap[workflowNodeParameterLabelEnforced]
				if parentByReferenceIdExists {
					inputsByReferenceId[referencId][workflowNodeParameterLabelEnforced] = parentByReferenceId
					outputsByReferenceId[referencId][workflowNodeParameterLabelEnforced] = parentByReferenceId
				}
			}
		}
	}

	updateReferencIds := make(map[channelIdType]bool)

	switch workflowNode.Type {
	case workflow_helpers.WorkflowNodeDataSourceTorqChannels:
		var params TorqChannelsConfiguration
		err := json.Unmarshal([]byte(workflowNode.Parameters), &params)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var channelIds []int
		switch params.Source {
		case "all":
			channelIds, err = getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelAllChannels)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Obtaining allChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		case "event":
			channelIds, err = getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelEventChannels)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Obtaining eventChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		case "eventXorAll":
			channelIds, _ = getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelEventChannels)
			if len(channelIds) == 0 {
				channelIds, err = getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelAllChannels)
				if err != nil {
					return core.Inactive, errors.Wrapf(err, "Obtaining allChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
			}
		}

		err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelChannels, channelIds)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Adding All ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case workflow_helpers.WorkflowNodeSetVariable:
		//variableName := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableName).ValueString
		//stringVariableParameter := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueString)
		//if stringVariableParameter.ValueString != "" {
		//	outputs[variableName] = stringVariableParameter.ValueString
		//} else {
		//	outputs[variableName] = fmt.Sprintf("%d", getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueNumber).ValueNumber)
		//}
	case workflow_helpers.WorkflowNodeFilterOnVariable:
		//variableName := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableName).ValueString
		//stringVariableParameter := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueString)
		//stringValue := ""
		//if stringVariableParameter.ValueString != "" {
		//	stringValue = stringVariableParameter.ValueString
		//} else {
		//	stringValue = fmt.Sprintf("%d", getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueNumber).ValueNumber)
		//}
		//if inputs[variableName] == stringValue {
		//	activeOutputIndex = 0
		//} else {
		//	activeOutputIndex = 1
		//}
	case workflow_helpers.WorkflowNodeChannelBalanceEventFilter:
		linkedChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelChannels)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var params ChannelBalanceEventFilterConfiguration
		err = json.Unmarshal([]byte(workflowNode.Parameters), &params)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var filteredChannelIds []int
		events, err := getChannelBalanceEvents(inputs)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(events) == 0 {
			if !params.IgnoreWhenEventless {
				return core.Inactive, errors.Wrapf(err, "No event(s) to filter found for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			filteredChannelIds = linkedChannelIds
		} else {
			if params.FilterClauses.Filter.FuncName != "" || len(params.FilterClauses.Or) != 0 || len(params.FilterClauses.And) != 0 {
				filteredChannelIds = filterChannelBalanceEventChannelIds(params.FilterClauses, linkedChannelIds, events)
			} else {
				filteredChannelIds = linkedChannelIds
			}
		}

		err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelChannels, filteredChannelIds)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case workflow_helpers.WorkflowNodeChannelFilter:
		linkedChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelChannels)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var params FilterClauses
		err = json.Unmarshal([]byte(workflowNode.Parameters), &params)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var filteredChannelIds []int
		if params.Filter.FuncName != "" || len(params.Or) != 0 || len(params.And) != 0 {
			var linkedChannels []channels.ChannelBody
			torqNodeIds := cache.GetAllTorqNodeIds()
			for _, torqNodeId := range torqNodeIds {
				linkedChannelsByNode, err := channels.GetChannelsByIds(torqNodeId, linkedChannelIds)
				if err != nil {
					return core.Inactive, errors.Wrapf(err, "Getting the linked channels to filters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				linkedChannels = append(linkedChannels, linkedChannelsByNode...)
			}
			filteredChannelIds = FilterChannelBodyChannelIds(params, linkedChannels)
		} else {
			filteredChannelIds = linkedChannelIds
		}

		err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelChannels, filteredChannelIds)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case workflow_helpers.WorkflowNodeAddTag, workflow_helpers.WorkflowNodeRemoveTag:
		linkedChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelChannels)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		err = addOrRemoveTags(db, linkedChannelIds, workflowNode)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Adding or removing tags with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
		}
	case workflow_helpers.WorkflowNodeChannelPolicyConfigurator:
		linkedChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelChannels)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		for channelId, labelValueMap := range inputsByReferenceId {
			if !slices.Contains(linkedChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			var routingPolicySettings ChannelPolicyConfiguration
			routingPolicySettings, err = processRoutingPolicyConfigurator(channelId, inputsByReferenceId, workflowNode)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			marshalledChannelPolicyConfiguration, err := json.Marshal(routingPolicySettings)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			outputsByReferenceId[channelId][workflow_helpers.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)
			updateReferencIds[channelId] = true
		}

		err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelChannels, linkedChannelIds)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case workflow_helpers.WorkflowNodeChannelPolicyAutoRun:
		linkedChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelChannels)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		for channelId, labelValueMap := range inputsByReferenceId {
			if !slices.Contains(linkedChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			var routingPolicySettings ChannelPolicyConfiguration
			routingPolicySettings, err = processRoutingPolicyConfigurator(channelId, inputsByReferenceId, workflowNode)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			err = processRoutingPolicyRun(ctx, db, routingPolicySettings, workflowNode, reference, workflowTriggerNode.Type)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			marshalledChannelPolicyConfiguration, err := json.Marshal(routingPolicySettings)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			outputsByReferenceId[channelId][workflow_helpers.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)

			marshalledResponse, err := json.Marshal(true)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Response with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			// TODO FIXME create a more uniform status object
			outputsByReferenceId[channelId][workflow_helpers.WorkflowParameterLabelStatus] = string(marshalledResponse)
			updateReferencIds[channelId] = true
		}

		err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelChannels, linkedChannelIds)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case workflow_helpers.WorkflowNodeChannelPolicyRun:
		for channelId, labelValueMap := range inputsByReferenceId {
			routingPolicySettingsString, exists := labelValueMap[workflow_helpers.WorkflowParameterLabelRoutingPolicySettings]
			if !exists {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			var routingPolicySettings ChannelPolicyConfiguration
			err := json.Unmarshal([]byte(routingPolicySettingsString), &routingPolicySettings)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Unmarshalling Routing Policy Configuration with for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if routingPolicySettings.ChannelId != 0 {
				err = processRoutingPolicyRun(ctx, db, routingPolicySettings, workflowNode, reference, workflowTriggerNode.Type)
				if err != nil {
					return core.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				marshalledResponse, err := json.Marshal(true)
				if err != nil {
					return core.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Response for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				// TODO FIXME create a more uniform status object
				outputsByReferenceId[channelId][workflow_helpers.WorkflowParameterLabelStatus] = string(marshalledResponse)
				updateReferencIds[channelId] = true
			}
		}
	case workflow_helpers.WorkflowNodeRebalanceConfigurator:
		var rebalanceConfiguration RebalanceConfiguration
		err := json.Unmarshal([]byte(workflowNode.Parameters), &rebalanceConfiguration)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		incomingChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelIncomingChannels)
		if err != nil {
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels {
				return core.Inactive, errors.Wrapf(err, "Obtaining incomingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels && len(incomingChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No IncomingChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		outgoingChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelOutgoingChannels)
		if err != nil {
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels {
				return core.Inactive, errors.Wrapf(err, "Obtaining outgoingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels && len(outgoingChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No OutgoingChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		unfocusedPath := createUnfocusedPath(workflowNode, rebalanceConfiguration, workflowNodes)

		for channelId, labelValueMap := range inputsByReferenceId {
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels && !slices.Contains(incomingChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels && !slices.Contains(outgoingChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			rebalanceConfiguration, err = processRebalanceConfigurator(rebalanceConfiguration, channelId, incomingChannelIds, outgoingChannelIds, inputsByReferenceId, workflowNode)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Processing Rebalance configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			rebalanceConfiguration.WorkflowUnfocusedPath = unfocusedPath

			marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			outputsByReferenceId[channelId][workflow_helpers.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
			updateReferencIds[channelId] = true
		}

		if incomingChannelIds != nil {
			err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelIncomingChannels, incomingChannelIds)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Adding Incoming ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if outgoingChannelIds != nil {
			err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelOutgoingChannels, outgoingChannelIds)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Adding Outgoing ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}
	case workflow_helpers.WorkflowNodeRebalanceAutoRun:
		var rebalanceConfiguration RebalanceConfiguration
		err := json.Unmarshal([]byte(workflowNode.Parameters), &rebalanceConfiguration)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		incomingChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelIncomingChannels)
		if err != nil {
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels {
				return core.Inactive, errors.Wrapf(err, "Obtaining incomingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels && len(incomingChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No IncomingChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		outgoingChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelOutgoingChannels)
		if err != nil {
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels {
				return core.Inactive, errors.Wrapf(err, "Obtaining outgoingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels && len(outgoingChannelIds) == 0 {
			return core.Inactive, errors.Wrapf(err, "No OutgoingChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		unfocusedPath := createUnfocusedPath(workflowNode, rebalanceConfiguration, workflowNodes)

		var rebalanceConfigurations []RebalanceConfiguration
		for channelId, labelValueMap := range inputsByReferenceId {
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels && !slices.Contains(incomingChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels && !slices.Contains(outgoingChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			rebalanceConfiguration, err = processRebalanceConfigurator(rebalanceConfiguration, channelId, incomingChannelIds, outgoingChannelIds, inputsByReferenceId, workflowNode)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Processing Rebalance configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if rebalanceConfiguration.AmountMsat == nil {
				continue
			}

			rebalanceConfiguration.WorkflowUnfocusedPath = unfocusedPath

			if len(rebalanceConfiguration.IncomingChannelIds) != 0 && len(rebalanceConfiguration.OutgoingChannelIds) != 0 {
				marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
				if err != nil {
					return core.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				outputsByReferenceId[channelId][workflow_helpers.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
				updateReferencIds[channelId] = true
				rebalanceConfigurations = append(rebalanceConfigurations, rebalanceConfiguration)
			}
		}

		eventChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelEventChannels)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Obtaining eventChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var responses []lightning_helpers.RebalanceResponse
		responses, err = processRebalanceRun(ctx, db, eventChannelIds, rebalanceConfigurations, workflowNode, reference)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Processing Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(rebalanceConfigurations) != 0 {
			marshalledResponses, err := json.Marshal(responses)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Marshalling Rebalance Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			// TODO FIXME create a more uniform status object
			outputs[workflow_helpers.WorkflowParameterLabelStatus] = string(marshalledResponses)
		}

		if incomingChannelIds != nil {
			err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelIncomingChannels, incomingChannelIds)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Adding Incoming ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if outgoingChannelIds != nil {
			err = setChannelIds(outputs, workflow_helpers.WorkflowParameterLabelOutgoingChannels, outgoingChannelIds)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Adding Outgoing ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}
	case workflow_helpers.WorkflowNodeRebalanceRun:
		var rebalanceConfigurations []RebalanceConfiguration
		for channelId, labelValueMap := range inputsByReferenceId {
			rebalanceConfigurationString, exists := labelValueMap[workflow_helpers.WorkflowParameterLabelRebalanceSettings]
			if !exists {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			var rebalanceConfiguration RebalanceConfiguration
			err := json.Unmarshal([]byte(rebalanceConfigurationString), &rebalanceConfiguration)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Unmarshalling Rebalance Configuration with for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if len(rebalanceConfiguration.IncomingChannelIds) != 0 && len(rebalanceConfiguration.OutgoingChannelIds) != 0 {
				marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
				if err != nil {
					return core.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				outputsByReferenceId[channelId][workflow_helpers.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
				updateReferencIds[channelId] = true
				rebalanceConfigurations = append(rebalanceConfigurations, rebalanceConfiguration)
			}
		}

		eventChannelIds, err := getChannelIds(inputs, workflow_helpers.WorkflowParameterLabelEventChannels)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Obtaining eventChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		responses, err := processRebalanceRun(ctx, db, eventChannelIds, rebalanceConfigurations, workflowNode, reference)
		if err != nil {
			return core.Inactive, errors.Wrapf(err, "Processing Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(rebalanceConfigurations) != 0 {
			marshalledResponses, err := json.Marshal(responses)
			if err != nil {
				return core.Inactive, errors.Wrapf(err, "Marshalling Rebalance Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			// TODO FIXME create a more uniform status object
			outputs[workflow_helpers.WorkflowParameterLabelStatus] = string(marshalledResponses)
		}
	}
	workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = core.Active

	if workflowStageOutputCache[stageType(workflowNode.Stage)] == nil {
		workflowStageOutputCache[stageType(workflowNode.Stage)] = make(map[workflow_helpers.WorkflowParameterLabel]string)
	}
	for label, value := range outputs {
		workflowStageOutputCache[stageType(workflowNode.Stage)][label] = value
	}

	if workflowStageOutputByReferenceIdCache[stageType(workflowNode.Stage)] == nil {
		workflowStageOutputByReferenceIdCache[stageType(workflowNode.Stage)] = make(map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string)
	}
	for channelId, labelValueMap := range outputsByReferenceId {
		if workflowStageOutputByReferenceIdCache[stageType(workflowNode.Stage)][channelId] == nil {
			workflowStageOutputByReferenceIdCache[stageType(workflowNode.Stage)][channelId] = make(map[workflow_helpers.WorkflowParameterLabel]string)
		}
		if updateReferencIds[channelId] {
			for label, value := range labelValueMap {
				workflowStageOutputByReferenceIdCache[stageType(workflowNode.Stage)][channelId][label] = value
			}
		}
	}

	marshalledInputs, err := json.Marshal([]any{inputs, inputsByReferenceId})
	if err != nil {
		log.Error().Err(err).Msgf("Marshalling inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	marshalledOutputs, err := json.Marshal([]any{outputs, outputsByReferenceId})
	if err != nil {
		log.Error().Err(err).Msgf("Marshalling outputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	_, err = addWorkflowVersionNodeLog(db, WorkflowVersionNodeLog{
		TriggerReference:                reference,
		InputData:                       string(marshalledInputs),
		OutputData:                      string(marshalledOutputs),
		DebugData:                       "",
		ErrorData:                       "",
		WorkflowVersionNodeId:           workflowNode.WorkflowVersionNodeId,
		TriggeringWorkflowVersionNodeId: &workflowTriggerNode.WorkflowVersionNodeId,
		CreatedOn:                       time.Now().UTC(),
	})
	if err != nil {
		log.Error().Err(err).Msgf("Storing log for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	return core.Active, nil
}

func createUnfocusedPath(
	workflowNode WorkflowNode,
	rebalanceConfiguration RebalanceConfiguration,
	workflowNodes []WorkflowNode) []WorkflowNode {

	var reversedPath []WorkflowNode
	childWorkflowNode := workflowNode
	goodPath := false
	for {
		var labelsOrdered []workflow_helpers.WorkflowParameterLabel
		switch childWorkflowNode.Type {
		case workflow_helpers.WorkflowNodeRebalanceAutoRun, workflow_helpers.WorkflowNodeRebalanceConfigurator:
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels {
				labelsOrdered = []workflow_helpers.WorkflowParameterLabel{
					workflow_helpers.WorkflowParameterLabelIncomingChannels,
					workflow_helpers.WorkflowParameterLabelOutgoingChannels,
				}
			}
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels {
				labelsOrdered = []workflow_helpers.WorkflowParameterLabel{
					workflow_helpers.WorkflowParameterLabelOutgoingChannels,
					workflow_helpers.WorkflowParameterLabelIncomingChannels,
				}
			}
		default:
			labelsOrdered = []workflow_helpers.WorkflowParameterLabel{
				workflow_helpers.WorkflowParameterLabelChannels,
			}
		}
		label, parentWorkflowNode := getParent(childWorkflowNode, workflowNodes, labelsOrdered)
		if label == "" {
			break
		}

		childWorkflowNode = parentWorkflowNode

		if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels &&
			label == workflow_helpers.WorkflowParameterLabelIncomingChannels {
			goodPath = true
		} else if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels &&
			label == workflow_helpers.WorkflowParameterLabelOutgoingChannels {
			goodPath = true
		} else if !goodPath {
			continue
		}
		reversedPath = append(reversedPath, WorkflowNode{
			WorkflowVersionNodeId: parentWorkflowNode.WorkflowVersionNodeId,
			Name:                  parentWorkflowNode.Name,
			Status:                parentWorkflowNode.Status,
			Stage:                 parentWorkflowNode.Stage,
			Type:                  parentWorkflowNode.Type,
			Parameters:            parentWorkflowNode.Parameters,
			WorkflowVersionId:     parentWorkflowNode.WorkflowVersionId,
		})
	}

	// change direction
	for i, j := 0, len(reversedPath)-1; i < j; i, j = i+1, j-1 {
		reversedPath[i], reversedPath[j] = reversedPath[j], reversedPath[i]
	}
	return reversedPath
}

func getParent(
	workflowNode WorkflowNode,
	workflowNodes []WorkflowNode,
	labelsOrdered []workflow_helpers.WorkflowParameterLabel) (workflow_helpers.WorkflowParameterLabel, WorkflowNode) {

	for _, label := range labelsOrdered {
		for _, link := range workflowNode.LinkDetails {
			if link.ChildWorkflowVersionNodeId == workflowNode.WorkflowVersionNodeId &&
				link.ChildInput == label {
				for _, parentWorkflowNode := range workflowNodes {
					if parentWorkflowNode.WorkflowVersionNodeId == link.ParentWorkflowVersionNodeId {
						return label, parentWorkflowNode
					}
				}
			}
		}
	}
	return "", WorkflowNode{}
}

func getChannelIds(inputs map[workflow_helpers.WorkflowParameterLabel]string, label workflow_helpers.WorkflowParameterLabel) ([]int, error) {
	channelIdsString, exists := inputs[label]
	if !exists {
		return nil, errors.New(fmt.Sprintf("Parse %v", label))
	}
	var channelIds []int
	err := json.Unmarshal([]byte(channelIdsString), &channelIds)
	if err != nil {
		return nil, errors.Wrapf(err, "Unmarshalling  %v", label)
	}
	return channelIds, nil
}

func getChannelBalanceEvents(inputs map[workflow_helpers.WorkflowParameterLabel]string) ([]core.ChannelBalanceEvent, error) {
	channelBalanceEventsString, exists := inputs[workflow_helpers.WorkflowParameterLabelEvents]
	if !exists {
		return nil, errors.New(fmt.Sprintf("Parse %v", workflow_helpers.WorkflowParameterLabelEvents))
	}
	var channelBalanceEvents []core.ChannelBalanceEvent
	err := json.Unmarshal([]byte(channelBalanceEventsString), &channelBalanceEvents)
	if err != nil {
		return nil, errors.Wrapf(err, "Unmarshalling  %v", workflow_helpers.WorkflowParameterLabelEvents)
	}
	if len(channelBalanceEvents) == 1 && channelBalanceEvents[0].ChannelId == 0 {
		return nil, nil
	}
	return channelBalanceEvents, nil
}

func setChannelIds(outputs map[workflow_helpers.WorkflowParameterLabel]string, label workflow_helpers.WorkflowParameterLabel, channelIds []int) error {
	ba, err := json.Marshal(channelIds)
	if err != nil {
		return errors.Wrapf(err, "Marshal the channelIds: %v", channelIds)
	}
	outputs[label] = string(ba)
	return nil
}

func processRebalanceConfigurator(
	rebalanceConfiguration RebalanceConfiguration,
	channelId channelIdType,
	incomingChannelIds []int,
	outgoingChannelIds []int,
	inputsByReferenceId map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowNode WorkflowNode) (RebalanceConfiguration, error) {

	var rebalanceInputConfiguration RebalanceConfiguration
	rebalanceInputConfigurationString, exists := inputsByReferenceId[channelId][workflow_helpers.WorkflowParameterLabelRebalanceSettings]
	if exists && rebalanceInputConfigurationString != "" && rebalanceInputConfigurationString != "null" {
		err := json.Unmarshal([]byte(rebalanceInputConfigurationString), &rebalanceInputConfiguration)
		if err != nil {
			return RebalanceConfiguration{}, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	}

	if rebalanceInputConfiguration.Focus != "" && rebalanceInputConfiguration.Focus != rebalanceConfiguration.Focus {
		return RebalanceConfiguration{}, errors.New(fmt.Sprintf("RebalanceConfiguration has mismatching focus for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId))
	}
	rebalanceInputConfiguration.Focus = rebalanceConfiguration.Focus

	if rebalanceConfiguration.AmountMsat != nil {
		rebalanceInputConfiguration.AmountMsat = rebalanceConfiguration.AmountMsat
	}
	if rebalanceConfiguration.MaximumCostMilliMsat != nil {
		rebalanceInputConfiguration.MaximumCostMilliMsat = rebalanceConfiguration.MaximumCostMilliMsat
		rebalanceInputConfiguration.MaximumCostMsat = nil
	}
	if rebalanceConfiguration.MaximumCostMsat != nil {
		rebalanceInputConfiguration.MaximumCostMilliMsat = nil
		rebalanceInputConfiguration.MaximumCostMsat = rebalanceConfiguration.MaximumCostMsat
	}
	if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels {
		rebalanceInputConfiguration.IncomingChannelIds = []int{int(channelId)}
		if len(outgoingChannelIds) != 0 {
			rebalanceInputConfiguration.OutgoingChannelIds = outgoingChannelIds
		}
	}
	if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels {
		if len(incomingChannelIds) != 0 {
			rebalanceInputConfiguration.IncomingChannelIds = incomingChannelIds
		}
		rebalanceInputConfiguration.OutgoingChannelIds = []int{int(channelId)}
	}
	return rebalanceInputConfiguration, nil
}

func processRebalanceRun(ctx context.Context,
	db *sqlx.DB,
	eventChannelIds []int,
	rebalanceSettings []RebalanceConfiguration,
	workflowNode WorkflowNode,
	reference string) ([]lightning_helpers.RebalanceResponse, error) {

	requestsMap := make(map[int]*lightning_helpers.RebalanceRequests)
	for _, rebalanceSetting := range rebalanceSettings {
		var maxCostMsat uint64
		if rebalanceSetting.MaximumCostMsat != nil {
			maxCostMsat = *rebalanceSetting.MaximumCostMsat
		}
		if rebalanceSetting.MaximumCostMilliMsat != nil {
			maxCostMsat = uint64(*rebalanceSetting.MaximumCostMilliMsat) * (*rebalanceSetting.AmountMsat) / 1_000_000
		}

		workflowUnfocusedPathMarshalled, err := json.Marshal(rebalanceSetting.WorkflowUnfocusedPath)
		if err != nil {
			return nil, errors.Wrapf(err,
				"Marshalling WorkflowUnfocusedPath of the rebalanceSetting for incomingChannelIds: %v, outgoingChannelIds: %v",
				rebalanceSetting.IncomingChannelIds, rebalanceSetting.OutgoingChannelIds)
		}

		if len(rebalanceSetting.WorkflowUnfocusedPath) != 0 {
			if rebalanceSetting.Focus == RebalancerFocusIncomingChannels {
				for _, incomingChannelId := range rebalanceSetting.IncomingChannelIds {
					var channelIds []int
					for _, outgoingChannelId := range rebalanceSetting.OutgoingChannelIds {
						if outgoingChannelId != 0 {
							channelIds = append(channelIds, outgoingChannelId)
						}
					}
					if len(channelIds) > 0 {
						channelSetting := cache.GetChannelSettingByChannelId(incomingChannelId)
						nodeId := channelSetting.FirstNodeId
						if !slices.Contains(cache.GetAllTorqNodeIds(), nodeId) {
							nodeId = channelSetting.SecondNodeId
						}
						_, exists := requestsMap[nodeId]
						if !exists {
							requestsMap[nodeId] = &lightning_helpers.RebalanceRequests{
								CommunicationRequest: lightning_helpers.CommunicationRequest{
									NodeId: nodeId,
								},
							}
						}
						requestsMap[nodeId].Requests = append(requestsMap[nodeId].Requests, lightning_helpers.RebalanceRequest{
							Origin:                lightning_helpers.RebalanceWorkflowNode,
							OriginId:              workflowNode.WorkflowVersionNodeId,
							OriginReference:       reference,
							IncomingChannelId:     incomingChannelId,
							OutgoingChannelId:     0,
							ChannelIds:            channelIds,
							AmountMsat:            *rebalanceSetting.AmountMsat,
							MaximumCostMsat:       maxCostMsat,
							MaximumConcurrency:    1,
							WorkflowUnfocusedPath: string(workflowUnfocusedPathMarshalled),
						})
					}
				}
			}
			if rebalanceSetting.Focus == RebalancerFocusOutgoingChannels {
				for _, outgoingChannelId := range rebalanceSetting.OutgoingChannelIds {
					var channelIds []int
					for _, incomingChannelId := range rebalanceSetting.IncomingChannelIds {
						if incomingChannelId != 0 {
							channelIds = append(channelIds, incomingChannelId)
						}
					}
					if len(channelIds) > 0 {
						channelSetting := cache.GetChannelSettingByChannelId(outgoingChannelId)
						nodeId := channelSetting.FirstNodeId
						if !slices.Contains(cache.GetAllTorqNodeIds(), nodeId) {
							nodeId = channelSetting.SecondNodeId
						}
						_, exists := requestsMap[nodeId]
						if !exists {
							requestsMap[nodeId] = &lightning_helpers.RebalanceRequests{
								CommunicationRequest: lightning_helpers.CommunicationRequest{
									NodeId: nodeId,
								},
							}
						}
						requestsMap[nodeId].Requests = append(requestsMap[nodeId].Requests, lightning_helpers.RebalanceRequest{
							Origin:                lightning_helpers.RebalanceWorkflowNode,
							OriginId:              workflowNode.WorkflowVersionNodeId,
							OriginReference:       reference,
							IncomingChannelId:     0,
							OutgoingChannelId:     outgoingChannelId,
							ChannelIds:            channelIds,
							AmountMsat:            *rebalanceSetting.AmountMsat,
							MaximumCostMsat:       maxCostMsat,
							MaximumConcurrency:    1,
							WorkflowUnfocusedPath: string(workflowUnfocusedPathMarshalled),
						})
					}
				}
			}
		}
	}
	var activeChannelIds []int
	var responses []lightning_helpers.RebalanceResponse
	for nodeId, requests := range requestsMap {
		if cache.GetCurrentNodeServiceState(services_helpers.LndServiceRebalanceService, nodeId).Status != services_helpers.Active {
			return nil, errors.New(fmt.Sprintf("Rebalance service is not active for nodeId: %v", nodeId))
		}
		reqs := *requests
		for _, req := range reqs.Requests {
			if req.IncomingChannelId != 0 {
				activeChannelIds = append(activeChannelIds, req.IncomingChannelId)
			}
			if req.OutgoingChannelId != 0 {
				activeChannelIds = append(activeChannelIds, req.OutgoingChannelId)
			}
		}
		resp := RebalanceRequests(ctx, db, reqs, nodeId)
		responses = append(responses, resp...)
	}
	if len(eventChannelIds) == 0 {
		CancelRebalancersExcept(lightning_helpers.RebalanceWorkflowNode, workflowNode.WorkflowVersionNodeId,
			activeChannelIds)
	} else {
		for _, eventChannelId := range eventChannelIds {
			if !slices.Contains(activeChannelIds, eventChannelId) {
				CancelRebalancer(lightning_helpers.RebalanceWorkflowNode, workflowNode.WorkflowVersionNodeId,
					eventChannelId)
			}
		}
	}
	return responses, nil
}

func processRoutingPolicyConfigurator(
	channelId channelIdType,
	inputsByChannelId map[channelIdType]map[workflow_helpers.WorkflowParameterLabel]string,
	workflowNode WorkflowNode) (ChannelPolicyConfiguration, error) {

	var channelPolicyInputConfiguration ChannelPolicyConfiguration
	channelPolicyInputConfigurationString, exists := inputsByChannelId[channelId][workflow_helpers.WorkflowParameterLabelRoutingPolicySettings]
	if exists && channelPolicyInputConfigurationString != "" && channelPolicyInputConfigurationString != "null" {
		err := json.Unmarshal([]byte(channelPolicyInputConfigurationString), &channelPolicyInputConfiguration)
		if err != nil {
			return ChannelPolicyConfiguration{}, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	}

	var channelPolicyConfiguration ChannelPolicyConfiguration
	err := json.Unmarshal([]byte(workflowNode.Parameters), &channelPolicyConfiguration)
	if err != nil {
		return ChannelPolicyConfiguration{}, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	if channelPolicyConfiguration.FeeBaseMsat != nil {
		channelPolicyInputConfiguration.FeeBaseMsat = channelPolicyConfiguration.FeeBaseMsat
	}
	if channelPolicyConfiguration.FeeRateMilliMsat != nil {
		channelPolicyInputConfiguration.FeeRateMilliMsat = channelPolicyConfiguration.FeeRateMilliMsat
	}
	if channelPolicyConfiguration.MaxHtlcMsat != nil {
		channelPolicyInputConfiguration.MaxHtlcMsat = channelPolicyConfiguration.MaxHtlcMsat
	}
	if channelPolicyConfiguration.MinHtlcMsat != nil {
		channelPolicyInputConfiguration.MinHtlcMsat = channelPolicyConfiguration.MinHtlcMsat
	}
	if channelPolicyConfiguration.TimeLockDelta != nil {
		channelPolicyInputConfiguration.TimeLockDelta = channelPolicyConfiguration.TimeLockDelta
	}
	channelPolicyInputConfiguration.ChannelId = int(channelId)
	return channelPolicyInputConfiguration, nil
}

func processRoutingPolicyRun(ctx context.Context,
	db *sqlx.DB,
	routingPolicySettings ChannelPolicyConfiguration,
	workflowNode WorkflowNode,
	reference string,
	triggerType workflow_helpers.WorkflowNodeType) error {

	torqNodeIds := cache.GetAllTorqNodeIds()
	channelSettings := cache.GetChannelSettingByChannelId(routingPolicySettings.ChannelId)
	nodeId := channelSettings.FirstNodeId
	if !slices.Contains(torqNodeIds, nodeId) {
		nodeId = channelSettings.SecondNodeId
	}
	if !slices.Contains(torqNodeIds, nodeId) {
		return errors.New(fmt.Sprintf("Routing policy update on unmanaged channel for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId))
	}
	rateLimitSeconds := 0
	rateLimitCount := 0
	if triggerType == workflow_helpers.WorkflowNodeManualTrigger {
		// DISABLE rate limiter
		rateLimitSeconds = 1
		rateLimitCount = 10
	}

	request := lightning_helpers.RoutingPolicyUpdateRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		Db:               db,
		RateLimitSeconds: rateLimitSeconds,
		RateLimitCount:   rateLimitCount,
		ChannelId:        routingPolicySettings.ChannelId,
		FeeRateMilliMsat: routingPolicySettings.FeeRateMilliMsat,
		FeeBaseMsat:      routingPolicySettings.FeeBaseMsat,
		MaxHtlcMsat:      routingPolicySettings.MaxHtlcMsat,
		MinHtlcMsat:      routingPolicySettings.MinHtlcMsat,
		TimeLockDelta:    routingPolicySettings.TimeLockDelta,
	}

	_, err := lightning.SetRoutingPolicy(ctx, request)
	if err != nil {
		log.Error().Err(err).Msgf("Workflow Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	return nil
}

func addOrRemoveTags(db *sqlx.DB, linkedChannelIds []int, workflowNode WorkflowNode) error {
	var params TagParameters
	err := json.Unmarshal([]byte(workflowNode.Parameters), &params)
	if err != nil {
		return errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}

	torqNodeIds := cache.GetAllTorqNodeIds()
	var processedNodeIds []int
	for _, tagToDelete := range params.RemovedTags {
		for index := range linkedChannelIds {
			var tag tags.TagEntityRequest
			processedNodeIds, tag = getTagEntityRequest(linkedChannelIds[index], tagToDelete.Value, params, torqNodeIds, processedNodeIds)
			if tag.TagId == 0 {
				continue
			}
			err = tags.UntagEntity(db, tag)
			if err != nil {
				return errors.Wrapf(err, "Failed to remove the tags for WorkflowVersionNodeId: %v tagIDd", workflowNode.WorkflowVersionNodeId, tagToDelete.Value)
			}
		}
	}

	processedNodeIds = []int{}
	for _, tagtoAdd := range params.AddedTags {
		for index := range linkedChannelIds {
			var tag tags.TagEntityRequest
			processedNodeIds, tag = getTagEntityRequest(linkedChannelIds[index], tagtoAdd.Value, params, torqNodeIds, processedNodeIds)
			if tag.TagId == 0 {
				continue
			}
			tag.CreatedByWorkflowVersionNodeId = &workflowNode.WorkflowVersionNodeId
			err = tags.TagEntity(db, tag)
			if err != nil {
				return errors.Wrapf(err, "Failed to add the tags for WorkflowVersionNodeId: %v tagIDd", workflowNode.WorkflowVersionNodeId, tagtoAdd.Value)
			}
		}
	}
	return nil
}

func getTagEntityRequest(channelId int, tagId int, params TagParameters, torqNodeIds []int, processedNodeIds []int) ([]int, tags.TagEntityRequest) {
	if params.ApplyTo == "nodes" {
		channelSettings := cache.GetChannelSettingByChannelId(channelId)
		nodeId := channelSettings.FirstNodeId
		if slices.Contains(torqNodeIds, nodeId) {
			nodeId = channelSettings.SecondNodeId
		}
		if slices.Contains(torqNodeIds, nodeId) {
			log.Info().Msgf("Both nodes are managed by Torq nodeIds: %v and %v", channelSettings.FirstNodeId, channelSettings.SecondNodeId)
			return processedNodeIds, tags.TagEntityRequest{}
		}
		if slices.Contains(processedNodeIds, nodeId) {
			return processedNodeIds, tags.TagEntityRequest{}
		}
		processedNodeIds = append(processedNodeIds, nodeId)
		return processedNodeIds, tags.TagEntityRequest{
			NodeId: &nodeId,
			TagId:  tagId,
		}
	} else {
		return processedNodeIds, tags.TagEntityRequest{
			ChannelId: &channelId,
			TagId:     tagId,
		}
	}
}

func filterChannelBalanceEventChannelIds(params FilterClauses, linkedChannelIds []int, events []core.ChannelBalanceEvent) []int {
	filteredChannelIds := extractChannelIds(ApplyFilters(params, ChannelBalanceEventToMap(events)))
	var resultChannelIds []int
	for _, linkedChannelId := range linkedChannelIds {
		if slices.Contains(filteredChannelIds, linkedChannelId) {
			resultChannelIds = append(resultChannelIds, linkedChannelId)
		}
	}
	return resultChannelIds
}

func FilterChannelBodyChannelIds(params FilterClauses, linkedChannels []channels.ChannelBody) []int {
	filteredChannelIds := extractChannelIds(ApplyFilters(params, ChannelBodyToMap(linkedChannels)))
	log.Trace().Msgf("Filtering applied to %d of %d channels", len(filteredChannelIds), len(linkedChannels))
	return filteredChannelIds
}

func extractChannelIds(filteredChannels []interface{}) []int {
	var filteredChannelIds []int
	for _, filteredChannel := range filteredChannels {
		channel, ok := filteredChannel.(map[string]interface{})
		if ok {
			filteredChannelIds = append(filteredChannelIds, channel["channelid"].(int))
			log.Trace().Msgf("Filter applied to channelId: %v", channel["channelid"])
		}
	}
	return filteredChannelIds
}

func AddWorkflowVersionNodeLog(db *sqlx.DB,
	reference string,
	workflowVersionNodeId int,
	triggeringWorkflowVersionNodeId int,
	inputs map[workflow_helpers.WorkflowParameterLabel]string,
	outputs map[workflow_helpers.WorkflowParameterLabel]string,
	workflowError error) {

	workflowVersionNodeLog := WorkflowVersionNodeLog{
		WorkflowVersionNodeId: workflowVersionNodeId,
		TriggerReference:      reference,
	}
	if triggeringWorkflowVersionNodeId > 0 {
		workflowVersionNodeLog.TriggeringWorkflowVersionNodeId = &triggeringWorkflowVersionNodeId
	}
	workflowVersionNodeLog.InputData = "[]"
	if len(inputs) > 0 {
		marshalledInputs, err := json.Marshal(inputs)
		if err == nil {
			workflowVersionNodeLog.InputData = string(marshalledInputs)
		} else {
			log.Error().Err(err).Msgf("Failed to marshal inputs for WorkflowVersionNodeId: %v", workflowVersionNodeId)
		}
	}
	workflowVersionNodeLog.OutputData = "[]"
	if len(outputs) != 0 {
		marshalledOutputs, err := json.Marshal(outputs)
		if err == nil {
			workflowVersionNodeLog.OutputData = string(marshalledOutputs)
		} else {
			log.Error().Err(err).Msgf("Failed to marshal outputs for WorkflowVersionNodeId: %v", workflowVersionNodeId)
		}
	}
	if workflowError != nil {
		workflowVersionNodeLog.ErrorData = workflowError.Error()
	}
	_, err := addWorkflowVersionNodeLog(db, workflowVersionNodeLog)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to log root node execution for workflowVersionNodeId: %v", workflowVersionNodeId)
	}
}

func ChannelBalanceEventToMap(structs []core.ChannelBalanceEvent) []map[string]interface{} {
	var maps []map[string]interface{}
	for _, s := range structs {
		maps = AddStructToMap(maps, s)
	}
	return maps
}

func ChannelBodyToMap(structs []channels.ChannelBody) []map[string]interface{} {
	var maps []map[string]interface{}
	for _, s := range structs {
		maps = AddStructToMap(maps, s)
	}
	return maps
}

func AddStructToMap(maps []map[string]interface{}, data any) []map[string]interface{} {
	structValue := reflect.ValueOf(data)
	structType := reflect.TypeOf(data)
	mapValue := make(map[string]interface{})

	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		mapValue[strings.ToLower(field.Name)] = structValue.Field(i).Interface()
	}
	maps = append(maps, mapValue)
	return maps
}
