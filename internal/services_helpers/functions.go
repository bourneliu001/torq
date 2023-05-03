package services_helpers

import (
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/core"
)

func (s *ServiceStatus) String() string {
	if s == nil {
		return core.UnknownEnumString
	}
	switch *s {
	case Inactive:
		return "Inactive"
	case Active:
		return "Active"
	case Pending:
		return "Pending"
	case Initializing:
		return "Initializing"
	}
	return core.UnknownEnumString
}

func GetCoreServiceTypes() []ServiceType {
	return []ServiceType{
		RootService,
		MaintenanceService,
		AutomationIntervalTriggerService,
		AutomationChannelBalanceEventTriggerService,
		AutomationChannelEventTriggerService,
		AutomationScheduledTriggerService,
		CronService,
		NotifierService,
		SlackService,
		TelegramHighService,
		TelegramLowService,
	}
}

func GetLndServiceTypes() []ServiceType {
	return []ServiceType{
		LndServiceVectorService,
		LndServiceAmbossService,
		LndServiceRebalanceService,
		LndServiceChannelEventStream,
		LndServiceGraphEventStream,
		LndServiceTransactionStream,
		LndServiceHtlcEventStream,
		LndServiceForwardsService,
		LndServiceInvoiceStream,
		LndServicePaymentsService,
		LndServicePeerEventStream,
		LndServiceInFlightPaymentsService,
		LndServiceChannelBalanceCacheService,
	}
}

func GetClnServiceTypes() []ServiceType {
	return []ServiceType{
		ClnServiceVectorService,
		ClnServiceAmbossService,
		ClnServicePeersService,
		ClnServiceChannelsService,
		ClnServiceClosedChannelsService,
		ClnServiceFundsService,
		ClnServiceNodesService,
		ClnServiceTransactionsService,
		ClnServiceForwardsService,
		ClnServiceInvoicesService,
		ClnServicePaymentsService,
	}
}

func (st *ServiceType) String() string {
	if st == nil {
		return core.UnknownEnumString
	}
	switch *st {
	case LndServiceVectorService:
		return "LndServiceVectorService"
	case LndServiceAmbossService:
		return "LndServiceAmbossService"
	case RootService:
		return "RootService"
	case AutomationChannelBalanceEventTriggerService:
		return "AutomationChannelBalanceEventTriggerService"
	case AutomationChannelEventTriggerService:
		return "AutomationChannelEventTriggerService"
	case AutomationIntervalTriggerService:
		return "AutomationIntervalTriggerService"
	case AutomationScheduledTriggerService:
		return "AutomationScheduledTriggerService"
	case LndServiceRebalanceService:
		return "LndServiceRebalanceService"
	case MaintenanceService:
		return "MaintenanceService"
	case CronService:
		return "CronService"
	case NotifierService:
		return "NotifierService"
	case SlackService:
		return "SlackService"
	case TelegramHighService:
		return "TelegramHighService"
	case TelegramLowService:
		return "TelegramLowService"
	case LndServiceChannelEventStream:
		return "LndServiceChannelEventStream"
	case LndServiceGraphEventStream:
		return "LndServiceGraphEventStream"
	case LndServiceTransactionStream:
		return "LndServiceTransactionStream"
	case LndServiceHtlcEventStream:
		return "LndServiceHtlcEventStream"
	case LndServiceForwardsService:
		return "LndServiceForwardsService"
	case LndServiceInvoiceStream:
		return "LndServiceInvoiceStream"
	case LndServicePaymentsService:
		return "LndServicePaymentsService"
	case LndServicePeerEventStream:
		return "LndServicePeerEventStream"
	case LndServiceInFlightPaymentsService:
		return "LndServiceInFlightPaymentsService"
	case LndServiceChannelBalanceCacheService:
		return "LndServiceChannelBalanceCacheService"
	case ClnServiceVectorService:
		return "ClnServiceVectorService"
	case ClnServiceAmbossService:
		return "ClnServiceAmbossService"
	case ClnServicePeersService:
		return "ClnServicePeersService"
	case ClnServiceChannelsService:
		return "ClnServiceChannelsService"
	case ClnServiceClosedChannelsService:
		return "ClnServiceClosedChannelsService"
	case ClnServiceFundsService:
		return "ClnServiceFundsService"
	case ClnServiceNodesService:
		return "ClnServiceNodesService"
	case ClnServiceTransactionsService:
		return "ClnServiceTransactionsService"
	case ClnServiceForwardsService:
		return "ClnServiceForwardsService"
	case ClnServiceInvoicesService:
		return "ClnServiceInvoicesService"
	case ClnServicePaymentsService:
		return "ClnServicePaymentsService"
	}
	return core.UnknownEnumString
}

func (st *ServiceType) IsChannelBalanceCache() bool {
	if st != nil && (*st == LndServiceForwardsService ||
		*st == LndServiceInvoiceStream ||
		*st == LndServicePaymentsService ||
		*st == LndServicePeerEventStream ||
		*st == LndServiceChannelEventStream ||
		*st == LndServiceGraphEventStream ||
		*st == ClnServicePeersService ||
		*st == ClnServiceChannelsService ||
		*st == ClnServiceFundsService) {
		return true
	}
	return false
}

func (st *ServiceType) IsLndService() bool {
	if st != nil && (*st == LndServiceVectorService ||
		*st == LndServiceAmbossService ||
		*st == LndServiceRebalanceService ||
		*st == LndServiceChannelEventStream ||
		*st == LndServiceGraphEventStream ||
		*st == LndServiceTransactionStream ||
		*st == LndServiceHtlcEventStream ||
		*st == LndServiceForwardsService ||
		*st == LndServiceInvoiceStream ||
		*st == LndServicePaymentsService ||
		*st == LndServicePeerEventStream ||
		*st == LndServiceInFlightPaymentsService ||
		*st == LndServiceChannelBalanceCacheService) {
		return true
	}
	return false
}

func (st *ServiceType) IsClnService() bool {
	if st != nil && (*st == ClnServiceVectorService ||
		*st == ClnServiceAmbossService ||
		*st == ClnServicePeersService ||
		*st == ClnServiceChannelsService ||
		*st == ClnServiceClosedChannelsService ||
		*st == ClnServiceFundsService ||
		*st == ClnServiceNodesService ||
		*st == ClnServiceTransactionsService ||
		*st == ClnServiceForwardsService ||
		*st == ClnServiceInvoicesService ||
		*st == ClnServicePaymentsService) {
		return true
	}
	return false
}

func (st *ServiceType) GetImplementation() *core.Implementation {
	if st == nil {
		return nil
	}
	if (*st).IsLndService() {
		lnd := core.LND
		return &lnd
	}
	if (*st).IsClnService() {
		cln := core.CLN
		return &cln
	}
	return nil
}

func (st *ServiceType) GetNodeConnectionDetailCustomSettings() []core.NodeConnectionDetailCustomSettings {
	if st == nil {
		return nil
	}
	switch *st {
	case LndServicePaymentsService, ClnServicePaymentsService:
		return []core.NodeConnectionDetailCustomSettings{core.ImportFailedPayments, core.ImportPayments}
	case LndServiceHtlcEventStream, ClnServiceHtlcsService:
		return []core.NodeConnectionDetailCustomSettings{core.ImportHtlcEvents}
	case LndServiceTransactionStream, ClnServiceTransactionsService:
		return []core.NodeConnectionDetailCustomSettings{core.ImportTransactions}
	case LndServiceInvoiceStream, ClnServiceInvoicesService:
		return []core.NodeConnectionDetailCustomSettings{core.ImportInvoices}
	case LndServiceForwardsService, ClnServiceForwardsService:
		return []core.NodeConnectionDetailCustomSettings{core.ImportForwards, core.ImportHistoricForwards}
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: ServiceType not supported")
		return nil
	}
}

func GetNodeConnectionDetailServiceType(implementation core.Implementation,
	cs core.NodeConnectionDetailCustomSettings) *ServiceType {
	switch {
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportFailedPayments),
		cs.HasNodeConnectionDetailCustomSettings(core.ImportPayments):
		p := LndServicePaymentsService
		if implementation == core.CLN {
			p = ClnServicePaymentsService
		}
		return &p
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportHtlcEvents):
		h := LndServiceHtlcEventStream
		if implementation == core.CLN {
			h = ClnServiceHtlcsService
		}
		return &h
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportTransactions):
		t := LndServiceTransactionStream
		if implementation == core.CLN {
			t = ClnServiceTransactionsService
		}
		return &t
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportInvoices):
		i := LndServiceInvoiceStream
		if implementation == core.CLN {
			i = ClnServiceInvoicesService
		}
		return &i
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportForwards),
		cs.HasNodeConnectionDetailCustomSettings(core.ImportHistoricForwards):
		f := LndServiceForwardsService
		if implementation == core.CLN {
			f = ClnServiceForwardsService
		}
		return &f
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: NodeConnectionDetailCustomSettings not supported")
		return nil
	}
}

func (st *ServiceType) GetPingSystem() *core.PingSystem {
	if st == nil {
		return nil
	}
	switch *st {
	case LndServiceAmbossService, ClnServiceAmbossService:
		amboss := core.Amboss
		return &amboss
	case LndServiceVectorService, ClnServiceVectorService:
		vector := core.Vector
		return &vector
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: ServiceType not supported")
		return nil
	}
}

func GetPingSystemServiceType(ps core.PingSystem) *ServiceType {
	switch {
	case ps.HasPingSystem(core.Vector):
		vectorService := LndServiceVectorService
		return &vectorService
	case ps.HasPingSystem(core.Amboss):
		ambossService := LndServiceAmbossService
		return &ambossService
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: PingSystem not supported")
		return nil
	}
}

func (i ImportType) String() string {
	switch i {
	case ImportChannelRoutingPolicies:
		return "ImportChannelRoutingPolicies"
	case ImportNodeInformation:
		return "ImportNodeInformation"
	case ImportAllChannels:
		return "ImportAllChannels"
	case ImportPendingChannels:
		return "ImportPendingChannels"
	case ImportPeerStatus:
		return "ImportPeerStatus"
	}
	return core.UnknownEnumString
}
