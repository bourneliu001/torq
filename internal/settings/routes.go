package settings

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/pkg/cln_connect"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/lncapital/torq/proto/cln"
)

type settings struct {
	SettingsId                      int        `json:"settingsId" db:"settings_id"`
	DefaultDateRange                string     `json:"defaultDateRange" db:"default_date_range"`
	DefaultLanguage                 string     `json:"defaultLanguage" db:"default_language"`
	PreferredTimezone               string     `json:"preferredTimezone" db:"preferred_timezone"`
	WeekStartsOn                    string     `json:"weekStartsOn" db:"week_starts_on"`
	TorqUuid                        string     `json:"torqUuid" db:"torq_uuid"`
	MixpanelOptOut                  bool       `json:"mixpanelOptOut" db:"mixpanel_opt_out"`
	SlackOAuthToken                 *string    `json:"slackOAuthToken" db:"slack_oauth_token"`
	SlackBotAppToken                *string    `json:"slackBotAppToken" db:"slack_bot_app_token"`
	TelegramHighPriorityCredentials *string    `json:"telegramHighPriorityCredentials" db:"telegram_high_priority_credentials"`
	TelegramLowPriorityCredentials  *string    `json:"telegramLowPriorityCredentials" db:"telegram_low_priority_credentials"`
	CreatedOn                       time.Time  `json:"createdOn" db:"created_on"`
	UpdateOn                        *time.Time `json:"updatedOn" db:"updated_on"`
}

type timeZone struct {
	Name string `json:"name" db:"name"`
}

type ConnectionDetails struct {
	NodeId            int
	Name              string
	GRPCAddress       string
	TLSFileBytes      []byte
	MacaroonFileBytes []byte
	Status            core.Status
	PingSystem        core.PingSystem
	CustomSettings    core.NodeConnectionDetailCustomSettings
}

func setAllLndServices(ctx context.Context,
	nodeId int,
	lndActive bool,
	customSettings core.NodeConnectionDetailCustomSettings,
	pingSystem core.PingSystem) bool {

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	if lndActive {
		return cache.ActivateLndService(ctxWithTimeout, nodeId, customSettings, pingSystem)
	}
	return cache.InactivateNodeService(ctxWithTimeout, nodeId)
}

func setService(ctx context.Context, serviceType services_helpers.ServiceType, nodeId int, active bool) bool {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	if active {
		return cache.ActivateNodeServiceState(ctxWithTimeout, serviceType, nodeId)
	}
	return cache.InactivateNodeServiceState(ctxWithTimeout, serviceType, nodeId)
}

func RegisterSettingRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getSettingsHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateSettingsHandler(c, db) })
	r.GET("nodeConnectionDetails", func(c *gin.Context) { getAllNodeConnectionDetailsHandler(c, db) })
	r.GET("nodeConnectionDetails/:nodeId", func(c *gin.Context) { getNodeConnectionDetailsHandler(c, db) })
	r.POST("nodeConnectionDetails", func(c *gin.Context) { addNodeConnectionDetailsHandler(c, db) })
	r.PUT("nodeConnectionDetails", func(c *gin.Context) { setNodeConnectionDetailsHandler(c, db) })
	r.PUT("nodeConnectionDetails/:nodeId/:statusId", func(c *gin.Context) {
		setNodeConnectionDetailsStatusHandler(c, db)
	})
	r.PUT("nodePingSystem/:nodeId/:pingSystem/:statusId", func(c *gin.Context) {
		setNodeConnectionDetailsPingSystemHandler(c, db)
	})
	r.PUT("nodeCustomSetting/:nodeId/:customSetting/:statusId", func(c *gin.Context) {
		setNodeConnectionDetailsCustomSettingHandler(c, db)
	})
	r.PUT("customSettings/:nodeId/:customSettings/:pingSystems", func(c *gin.Context) {
		setNodeConnectionDetailsCustomSettingsHandler(c, db)
	})
}
func RegisterUnauthenticatedRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("timezones", func(c *gin.Context) { getTimeZonesHandler(c, db) })
}

func getTimeZonesHandler(c *gin.Context, db *sqlx.DB) {
	timeZones, err := getTimeZones(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, timeZones)
}
func getSettingsHandler(c *gin.Context, db *sqlx.DB) {
	setts, err := getSettings(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, setts)
}

func updateSettingsHandler(c *gin.Context, db *sqlx.DB) {
	var setts settings
	if err := c.BindJSON(&setts); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	err := updateSettings(db, setts)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	if setts.SlackBotAppToken != nil && *setts.SlackBotAppToken != "" &&
		setts.SlackOAuthToken != nil && *setts.SlackOAuthToken != "" {
		cache.SetDesiredCoreServiceState(services_helpers.SlackService, services_helpers.Active)
		cache.SetDesiredCoreServiceState(services_helpers.NotifierService, services_helpers.Active)
	}
	if setts.TelegramHighPriorityCredentials != nil && *setts.TelegramHighPriorityCredentials != "" {
		cache.SetDesiredCoreServiceState(services_helpers.TelegramHighService, services_helpers.Active)
		cache.SetDesiredCoreServiceState(services_helpers.NotifierService, services_helpers.Active)
	}
	if setts.TelegramLowPriorityCredentials != nil && *setts.TelegramLowPriorityCredentials != "" {
		cache.SetDesiredCoreServiceState(services_helpers.TelegramLowService, services_helpers.Active)
		cache.SetDesiredCoreServiceState(services_helpers.NotifierService, services_helpers.Active)
	}
	c.JSON(http.StatusOK, setts)
}

func getAllNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	node, err := GetAllNodeConnectionDetails(db, false)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, node)
}

func getNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	ncd, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, ncd)
}

func addNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	var ncd NodeConnectionDetails
	var err error
	existingNcd := NodeConnectionDetails{}

	if err = c.Bind(&ncd); err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	ncd, err = fixBindFailures(c, ncd)
	if err != nil {
		server_errors.SendBadRequest(c, err.Error())
		return
	}

	var publicKey string
	var chain core.Chain
	var network core.Network
	var canSignMessages bool
	switch ncd.Implementation {
	case core.LND:
		if ncd.TLSFile == nil || ncd.MacaroonFile == nil || ncd.GRPCAddress == nil || *ncd.GRPCAddress == "" {
			server_errors.SendBadRequest(c, "All node details are required to add new node connection details")
			return
		}

		ncd.TLSFileName, ncd.TLSDataBytes, err = processFile(ncd.TLSFile)
		if err != nil {
			server_errors.LogAndSendServerError(c, errors.New("Reading TLS File"))
			return
		}
		if len(ncd.TLSDataBytes) == 0 {
			server_errors.SendBadRequest(c, "Can't check new gRPC details without TLS Cert")
			return
		}

		ncd.MacaroonFileName, ncd.MacaroonDataBytes, err = processFile(ncd.MacaroonFile)
		if err != nil {
			server_errors.LogAndSendServerError(c, errors.New("Reading Macaroon File"))
			return
		}
		if len(ncd.MacaroonDataBytes) == 0 {
			server_errors.SendBadRequest(c, "Can't check new gRPC details without Macaroon File")
			return
		}

		publicKey, chain, network, err =
			getInformationFromLndNode(c.Request.Context(), *ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
		if err != nil {
			server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get info from LND Node"), "connect", map[string]string{"implementation": "LND"})
			return
		}

		ncd.NodeStartDate, err =
			getNodeStartDateFromLndNode(c.Request.Context(), *ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
		if err != nil {
			server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get node start date from LND Node"), "connect", map[string]string{"implementation": "LND"})
			return
		}

		canSignMessages, err =
			getSignMessagesPermissionFromLndNode(c.Request.Context(), *ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Verify sign messages macaroon ability from gRPC")
			return
		}
		if !canSignMessages {
			ncd.PingSystem = 0
		}

	case core.CLN:
		if ncd.CaCertificateFile == nil || ncd.CertificateFile == nil || ncd.KeyFile == nil || ncd.GRPCAddress == nil || *ncd.GRPCAddress == "" {
			server_errors.SendBadRequest(c, "All node details are required to add new node connection details")
			return
		}

		ncd.CaCertificateFileName, ncd.CaCertificateDataBytes, err = processFile(ncd.CaCertificateFile)
		if err != nil {
			server_errors.LogAndSendServerError(c, errors.New("Reading CA Certificate File"))
			return
		}
		if len(ncd.CaCertificateDataBytes) == 0 {
			server_errors.SendBadRequest(c, "Can't check new gRPC details without TLS Cert")
			return
		}
		ncd.CertificateFileName, ncd.CertificateDataBytes, err = processFile(ncd.CertificateFile)
		if err != nil {
			server_errors.LogAndSendServerError(c, errors.New("Reading Client Certificate File"))
			return
		}
		if len(ncd.CertificateDataBytes) == 0 {
			server_errors.SendBadRequest(c, "Can't check new gRPC details without Client Certificate File")
			return
		}
		ncd.KeyFileName, ncd.KeyDataBytes, err = processFile(ncd.KeyFile)
		if err != nil {
			server_errors.LogAndSendServerError(c, errors.New("Reading Client Private Key File"))
			return
		}
		if len(ncd.KeyDataBytes) == 0 {
			server_errors.SendBadRequest(c, "Can't check new gRPC details without Client Private Key File")
			return
		}

		publicKey, chain, network, err =
			getInformationFromClnNode(c.Request.Context(), *ncd.GRPCAddress, ncd.CertificateDataBytes, ncd.KeyDataBytes, ncd.CaCertificateDataBytes)
		if err != nil {
			server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get info from CLN Node"), "connect", map[string]string{"implementation": "CLN"})
			return
		}

		//ncd.NodeStartDate, err = getNodeStartDateFromClnNode(*ncd.GRPCAddress, ncd.CaCertificateDataBytes, ncd.CertificateDataBytes, ncd.KeyDataBytes)
		// TODO FIXME CLN: Once the CLN gRPC support paging we can do this.
		//if err != nil {
		//	server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get node start date from CLN Node"), "connect", map[string]string{"implementation": "CLN"})
		//	return
		//}
	}

	nodeId := cache.GetPeerNodeIdByPublicKey(publicKey, chain, network)
	if nodeId == 0 {
		nodeId, err = AddNodeWhenNew(db, publicKey, chain, network)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Adding node")
			return
		}
	}
	ncd.NodeId = nodeId
	existingNcd, err = getNodeConnectionDetails(db, ncd.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting existing node connection details by nodeId")
		return
	}
	if existingNcd.NodeId == ncd.NodeId && existingNcd.Status != core.Deleted {
		server_errors.SendUnprocessableEntity(c, "This node already exists")
		return
	}
	if existingNcd.Status == core.Deleted {
		ncd.CreateOn = existingNcd.CreateOn
		if strings.TrimSpace(ncd.Name) == "" {
			ncd.Name = existingNcd.Name
		}
	}

	if strings.TrimSpace(ncd.Name) == "" {
		ncd.Name = fmt.Sprintf("Node_%v", ncd.NodeId)
	}
	ncd.Status = core.Active
	if existingNcd.NodeId == ncd.NodeId {
		ncd, err = SetNodeConnectionDetails(db, ncd)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Updating and reactivating deleted node connection details")
			return
		}
	} else {
		ncd, err = addNodeConnectionDetails(db, ncd)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Adding node connection details")
			return
		}
	}
	cache.SetTorqNode(ncd.NodeId, ncd.Name, ncd.Status, publicKey, chain, network)

	if ncd.Status == core.Active {
		ctxWithTimeout, cancel := context.WithTimeout(c.Request.Context(), 1*time.Minute)
		defer cancel()

		if !cache.ActivateLndService(ctxWithTimeout, nodeId, ncd.CustomSettings, ncd.PingSystem) {
			server_errors.WrapLogAndSendServerError(c, err, "Service activation failed.")
			return
		}
	}

	c.JSON(http.StatusOK, ncd)
}

func setNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	var ncd NodeConnectionDetails
	var err error
	if err = c.Bind(&ncd); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	ncd, err = fixBindFailures(c, ncd)
	if err != nil {
		server_errors.SendBadRequest(c, err.Error())
		return
	}

	if ncd.NodeId == 0 {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	if ncd.GRPCAddress == nil || *ncd.GRPCAddress == "" {
		c.JSON(http.StatusBadRequest, server_errors.SingleFieldErrorCode("grpcAddress", "missingField", map[string]string{"field": "GRPC Address"}))
		return
	}
	if strings.TrimSpace(ncd.Name) == "" {
		ncd.Name = fmt.Sprintf("Node_%v", ncd.NodeId)
	}

	switch ncd.Implementation {
	case core.LND:
		ncd.TLSFileName, ncd.TLSDataBytes, err = processFile(ncd.TLSFile)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Processing TLS file")
			return
		}
		ncd.MacaroonFileName, ncd.MacaroonDataBytes, err = processFile(ncd.MacaroonFile)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Processing Macaroon file")
			return
		}
	case core.CLN:
		ncd.CaCertificateFileName, ncd.CaCertificateDataBytes, err = processFile(ncd.CaCertificateFile)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Processing CA Certificate file")
			return
		}
		ncd.CertificateFileName, ncd.CertificateDataBytes, err = processFile(ncd.CertificateFile)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Processing Client Certificate file")
			return
		}
		ncd.KeyFileName, ncd.KeyDataBytes, err = processFile(ncd.KeyFile)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Processing Client Private Key file")
			return
		}
	}

	existingNcd, err := getNodeConnectionDetails(db, ncd.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Obtaining existing node connection details")
		return
	}

	detailsUpdate := false
	if len(ncd.MacaroonDataBytes) > 0 ||
		len(ncd.TLSDataBytes) > 0 ||
		len(ncd.CaCertificateDataBytes) > 0 ||
		len(ncd.CertificateDataBytes) > 0 ||
		len(ncd.KeyDataBytes) > 0 ||
		*ncd.GRPCAddress != *existingNcd.GRPCAddress {
		detailsUpdate = true
	}

	if len(ncd.TLSDataBytes) == 0 && len(existingNcd.TLSDataBytes) != 0 {
		ncd.TLSDataBytes = existingNcd.TLSDataBytes
		ncd.TLSFileName = existingNcd.TLSFileName
	}
	if len(ncd.TLSDataBytes) == 0 {
		server_errors.LogAndSendServerError(c, errors.New("Can't check new gRPC details without TLS Cert"))
		return
	}

	if len(ncd.MacaroonDataBytes) == 0 && len(existingNcd.MacaroonDataBytes) != 0 {
		ncd.MacaroonDataBytes = existingNcd.MacaroonDataBytes
		ncd.MacaroonFileName = existingNcd.MacaroonFileName
	}
	if len(ncd.MacaroonDataBytes) == 0 {
		server_errors.LogAndSendServerError(c, errors.New("Can't check new gRPC details without Macaroon File"))
		return
	}

	if detailsUpdate {
		switch ncd.Implementation {
		case core.LND:
			publicKey, chain, network, err :=
				getInformationFromLndNode(c.Request.Context(), *ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
			if err != nil {
				server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get info from LND Node"), "connect", map[string]string{"implementation": "LND"})
				return
			}

			existingPublicKey, existingChain, existingNetwork, err := GetNodeDetailsById(db, ncd.NodeId)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Obtaining existing node")
				return
			}
			if existingPublicKey != publicKey || existingChain != chain || existingNetwork != network {
				server_errors.SendUnprocessableEntity(c, "PublicKey/chain/network does not match, create a new node instead of updating this one")
				return
			}

			// Get node start date from LND if user didn't give it manually
			if ncd.NodeStartDate == nil {
				nodeStartDate, err :=
					getNodeStartDateFromLndNode(c.Request.Context(), *ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
				if err != nil {
					server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get node start date from LND Node"), "connect", map[string]string{"implementation": "LND"})
					return
				}
				ncd.NodeStartDate = nodeStartDate
			}

			canSignMessages, err :=
				getSignMessagesPermissionFromLndNode(c.Request.Context(), *ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Verify sign messages macaroon ability from gRPC")
				return
			}

			if !canSignMessages {
				ncd.PingSystem = 0
			}
		case core.CLN:
			publicKey, chain, network, err :=
				getInformationFromClnNode(c.Request.Context(), *ncd.GRPCAddress, ncd.CertificateDataBytes, ncd.KeyDataBytes, ncd.CaCertificateDataBytes)
			if err != nil {
				server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get info from CLN Node"), "connect", map[string]string{"implementation": "CLN"})
				return
			}

			existingPublicKey, existingChain, existingNetwork, err := GetNodeDetailsById(db, ncd.NodeId)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Obtaining existing node")
				return
			}
			if existingPublicKey != publicKey || existingChain != chain || existingNetwork != network {
				server_errors.SendUnprocessableEntity(c, "PublicKey/chain/network does not match, create a new node instead of updating this one")
				return
			}

			//// Get node start date from CLN if user didn't give it manually
			// TODO FIXME CLN: Once the CLN gRPC support paging we can do this.
			//if ncd.NodeStartDate == nil {
			//	nodeStartDate, err := getNodeStartDateFromClnNode(*ncd.GRPCAddress, ncd.CaCertificateDataBytes, ncd.CertificateDataBytes, ncd.KeyDataBytes)
			//	if err != nil {
			//		server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get node start date from LND Node"), "connect", map[string]string{"implementation": "LND"})
			//		return
			//	}
			//	ncd.NodeStartDate = nodeStartDate
			//}
		}

	}

	nodeSettings := cache.GetNodeSettingsByNodeId(ncd.NodeId)
	if ncd.PingSystem.HasPingSystem(core.Amboss) &&
		(nodeSettings.Chain != core.Bitcoin && nodeSettings.Network != core.MainNet) {
		server_errors.LogAndSendServerError(c, errors.New("Amboss Ping Service is only allowed on Bitcoin Mainnet."))
		return
	}
	if ncd.PingSystem.HasPingSystem(core.Vector) &&
		(nodeSettings.Chain != core.Bitcoin && nodeSettings.Network != core.MainNet) {
		server_errors.LogAndSendServerError(c, errors.New("Vector Ping Service is only allowed on Bitcoin Mainnet."))
		return
	}

	ncd, err = SetNodeConnectionDetails(db, ncd)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Updating connection details")
		return
	}
	cache.SetTorqNode(ncd.NodeId, ncd.Name, ncd.Status,
		nodeSettings.PublicKey, nodeSettings.Chain, nodeSettings.Network)

	success := setAllLndServices(c.Request.Context(), ncd.NodeId, ncd.Status == core.Active, ncd.CustomSettings, ncd.PingSystem)
	if !success {
		server_errors.LogAndSendServerError(c, errors.New("Some services did not start correctly"))
		return
	}

	c.JSON(http.StatusOK, ncd)
}

func fixBindFailures(c *gin.Context, ncd NodeConnectionDetails) (NodeConnectionDetails, error) {
	// TODO c.Bind cannot process status?
	implementation, err := strconv.Atoi(c.Request.Form.Get("implementation"))
	if err != nil {
		return NodeConnectionDetails{}, errors.New("Failed to find/parse implementation in the request.")
	}
	if implementation > int(core.CLN) {
		return NodeConnectionDetails{}, errors.New("Failed to parse implementation in the request.")
	}
	ncd.Implementation = core.Implementation(implementation)

	// TODO c.Bind cannot process status?
	statusId, err := strconv.Atoi(c.Request.Form.Get("status"))
	if err != nil {
		return NodeConnectionDetails{}, errors.New("Failed to find/parse status in the request.")
	}
	if statusId > int(core.Archived) {
		return NodeConnectionDetails{}, errors.New("Failed to parse status in the request.")
	}
	ncd.Status = core.Status(statusId)

	// TODO c.Bind cannot process pingSystem?
	pingSystem, err := strconv.Atoi(c.Request.Form.Get("pingSystem"))
	if err != nil {
		return NodeConnectionDetails{}, errors.New("Failed to find/parse pingSystem in the request.")
	}
	if pingSystem > core.PingSystemMax {
		return NodeConnectionDetails{}, errors.New("Failed to parse pingSystem in the request.")
	}
	ncd.PingSystem = core.PingSystem(pingSystem)

	// TODO c.Bind cannot process customSettings?
	customSettings, err := strconv.Atoi(c.Request.Form.Get("customSettings"))
	if err != nil {
		return NodeConnectionDetails{}, errors.New("Failed to find/parse customSettings in the request.")
	}
	if customSettings > core.NodeConnectionDetailCustomSettingsMax {
		return NodeConnectionDetails{}, errors.New("Failed to parse customSettings in the request.")
	}
	ncd.CustomSettings = core.NodeConnectionDetailCustomSettings(customSettings)

	nodeStartDateInput := c.Request.Form.Get("nodeStartDate")
	if nodeStartDateInput != "" {
		nodeStartDate, err := time.Parse("2006-01-02", nodeStartDateInput)
		if err != nil {
			return NodeConnectionDetails{}, errors.New("Failed to parse nodeStartDate in the request.")
		}
		ncd.NodeStartDate = &nodeStartDate
	} else {
		ncd.NodeStartDate = nil
	}

	nodeCssColour := c.Request.Form.Get("nodeCssColour")
	if nodeCssColour != "" {
		ncd.NodeCssColour = &nodeCssColour
	} else {
		ncd.NodeCssColour = nil
	}

	return ncd, nil
}

func setNodeConnectionDetailsStatusHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	statusId, err := strconv.Atoi(c.Param("statusId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse statusId in the request.")
		return
	}

	_, err = setNodeConnectionDetailsStatus(db, nodeId, core.Status(statusId))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	switch core.Status(statusId) {
	case core.Deleted:
		node := cache.GetNodeSettingsByNodeId(nodeId)
		cache.RemoveNodeFromCache(node)
		cache.RemoveChannelStatesFromCache(node.NodeId)
	default:
		nodeSettings := cache.GetNodeSettingsByNodeId(nodeId)
		var name string
		if nodeSettings.Name != nil {
			name = *nodeSettings.Name
		}
		cache.SetTorqNode(nodeId, name, core.Status(statusId),
			nodeSettings.PublicKey, nodeSettings.Chain, nodeSettings.Network)
	}

	if core.Status(statusId) != core.Active {
		success := setAllLndServices(c.Request.Context(), nodeId, false, 0, 0)
		if !success {
			server_errors.LogAndSendServerError(c, errors.New("Some services did not start correctly"))
			return
		}
		c.Status(http.StatusOK)
		return
	}

	ncd, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	success := setAllLndServices(c.Request.Context(), nodeId, true, ncd.CustomSettings, ncd.PingSystem)
	if !success {
		server_errors.LogAndSendServerError(c, errors.New("Some services did not start correctly"))
		return
	}
	c.Status(http.StatusOK)
}

func setNodeConnectionDetailsPingSystemHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	pingSystem, err := strconv.Atoi(c.Param("pingSystem"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse pingSystem in the request.")
		return
	}
	if pingSystem > core.PingSystemMax {
		server_errors.SendBadRequest(c, "Failed to parse pingSystem in the request.")
		return
	}
	ps := core.PingSystem(pingSystem)
	statusId, err := strconv.Atoi(c.Param("statusId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse statusId in the request.")
		return
	}
	nodeSettings := cache.GetNodeSettingsByNodeId(nodeId)
	if core.Status(statusId) == core.Active &&
		(nodeSettings.Chain != core.Bitcoin || nodeSettings.Network != core.MainNet) {
		server_errors.SendBadRequest(c, "Ping Services are only allowed on Bitcoin Mainnet.")
		return
	}
	s := core.Status(statusId)

	_, err = setNodeConnectionDetailsPingSystemStatus(db, nodeId, ps, s)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	done := setService(c.Request.Context(), *services_helpers.GetPingSystemServiceType(ps), nodeId, s == core.Active)
	if !done {
		server_errors.LogAndSendServerError(c, errors.New("Service failed please try again."))
		return
	}

	c.Status(http.StatusOK)
}

func setNodeConnectionDetailsCustomSettingHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	customSetting, err := strconv.Atoi(c.Param("customSetting"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse customSetting in the request.")
		return
	}
	if customSetting > core.NodeConnectionDetailCustomSettingsMax {
		server_errors.SendBadRequest(c, "Failed to parse customSetting in the request.")
		return
	}
	cs := core.NodeConnectionDetailCustomSettings(customSetting)
	statusId, err := strconv.Atoi(c.Param("statusId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse statusId in the request.")
		return
	}
	s := core.Status(statusId)

	_, err = setNodeConnectionDetailsCustomSettingStatus(db, nodeId, cs, s)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	ncd := cache.GetNodeConnectionDetails(nodeId)
	done := setService(c.Request.Context(),
		*services_helpers.GetNodeConnectionDetailServiceType(ncd.Implementation, cs), nodeId, s == core.Active)
	if !done {
		server_errors.LogAndSendServerError(c, errors.New("Service failed please try again."))
		return
	}

	c.Status(http.StatusOK)
}

func setNodeConnectionDetailsCustomSettingsHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	customSettings, err := strconv.Atoi(c.Param("customSettings"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse customSetting in the request.")
		return
	}
	if customSettings > core.NodeConnectionDetailCustomSettingsMax {
		server_errors.SendBadRequest(c, "Failed to parse customSetting in the request.")
		return
	}
	cs := core.NodeConnectionDetailCustomSettings(customSettings)

	pingSystems, err := strconv.Atoi(c.Param("pingSystems"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse statusId in the request.")
		return
	}
	if pingSystems > core.PingSystemMax {
		server_errors.SendBadRequest(c, "Failed to parse customSetting in the request.")
		return
	}
	ps := core.PingSystem(pingSystems)

	nodeSettings := cache.GetNodeSettingsByNodeId(nodeId)
	if (ps.HasPingSystem(core.Vector) || ps.HasPingSystem(core.Amboss)) &&
		(nodeSettings.Chain != core.Bitcoin || nodeSettings.Network != core.MainNet) {
		server_errors.SendBadRequest(c, "Ping Services are only allowed on Bitcoin Mainnet.")
		return
	}

	ncd := cache.GetNodeConnectionDetails(nodeId)
	services := make(map[bool][]services_helpers.ServiceType)
	appendPingSystemServiceType(ps, services, core.Vector)
	appendPingSystemServiceType(ps, services, core.Amboss)
	appendCustomSettingServiceType(ncd.Implementation, cs, services, core.ImportFailedPayments, core.ImportPayments)
	appendCustomSettingServiceType(ncd.Implementation, cs, services, core.ImportHtlcEvents)
	appendCustomSettingServiceType(ncd.Implementation, cs, services, core.ImportTransactions)
	appendCustomSettingServiceType(ncd.Implementation, cs, services, core.ImportInvoices)
	appendCustomSettingServiceType(ncd.Implementation, cs, services, core.ImportForwards, core.ImportHistoricForwards)

	_, err = setCustomSettings(db, nodeId, cs, ps)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	allSuccess := true
	for active, serviceTypes := range services {
		for _, serviceType := range serviceTypes {
			done := setService(c.Request.Context(), serviceType, nodeId, active)
			if !done {
				allSuccess = false
			}
		}
	}
	if !allSuccess {
		server_errors.LogAndSendServerError(c, errors.New("Service failed please try again."))
		return
	}

	c.Status(http.StatusOK)
}

func appendPingSystemServiceType(ps core.PingSystem,
	services map[bool][]services_helpers.ServiceType,
	referencePingSystem core.PingSystem) {

	if ps.HasPingSystem(referencePingSystem) {
		services[true] = append(services[true], *services_helpers.GetPingSystemServiceType(referencePingSystem))
	} else {
		services[false] = append(services[false], *services_helpers.GetPingSystemServiceType(referencePingSystem))
	}
}

func appendCustomSettingServiceType(implementation core.Implementation,
	cs core.NodeConnectionDetailCustomSettings,
	services map[bool][]services_helpers.ServiceType,
	referenceCustomSettings ...core.NodeConnectionDetailCustomSettings) {

	active := false
	var serviceType *services_helpers.ServiceType
	for ix, referenceCustomSetting := range referenceCustomSettings {
		st := services_helpers.GetNodeConnectionDetailServiceType(implementation, referenceCustomSettings[ix])
		if serviceType == nil {
			serviceType = st
		}
		if *st != *serviceType {
			return
		}
		if cs.HasNodeConnectionDetailCustomSettings(referenceCustomSetting) {
			active = true
		}
	}
	if active {
		services[true] = append(services[true], *serviceType)
	} else {
		services[false] = append(services[false], *serviceType)
	}
}

// GetConnectionDetailsById will still fetch details even if node is disabled or deleted
func GetConnectionDetailsById(db *sqlx.DB, nodeId int) (ConnectionDetails, error) {
	ncd, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		return ConnectionDetails{}, errors.Wrapf(err, "Getting node connection details from db for nodeId: %v", nodeId)
	}
	cd := ConnectionDetails{
		NodeId:            ncd.NodeId,
		TLSFileBytes:      ncd.TLSDataBytes,
		MacaroonFileBytes: ncd.MacaroonDataBytes,
		Name:              ncd.Name,
		Status:            ncd.Status,
		PingSystem:        ncd.PingSystem,
		CustomSettings:    ncd.CustomSettings,
	}
	if ncd.GRPCAddress != nil {
		cd.GRPCAddress = *ncd.GRPCAddress
	}
	return cd, nil
}

func getInformationFromLndNode(ctx context.Context, grpcAddress string, tlsCert []byte, macaroonFile []byte) (
	string, core.Chain, core.Network, error) {
	conn, err := lnd_connect.Connect(grpcAddress, tlsCert, macaroonFile)
	if err != nil {
		return "", 0, 0, errors.Wrap(err,
			"Can't connect to node to verify public key, check all details including TLS Cert and Macaroon")
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to close gRPC connection.")
		}
	}(conn)

	client := lnrpc.NewLightningClient(conn)
	info, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return "", 0, 0, errors.Wrap(err, "Obtaining information from LND")
	}
	if len(info.Chains) != 1 {
		return "", 0, 0, errors.Wrapf(err, "Obtaining chains from LND %v", info.Chains)
	}

	var chain core.Chain
	switch info.Chains[0].Chain {
	case "bitcoin":
		chain = core.Bitcoin
	case "litecoin":
		chain = core.Litecoin
	default:
		return "", 0, 0, errors.Wrapf(err, "Obtaining chain from LND %v", info.Chains[0].Chain)
	}

	var network core.Network
	switch info.Chains[0].Network {
	case "mainnet":
		network = core.MainNet
	case "testnet":
		network = core.TestNet
	case "signet":
		network = core.SigNet
	case "simnet":
		network = core.SimNet
	case "regtest":
		network = core.RegTest
	default:
		return "", 0, 0, errors.Wrapf(err, "Obtaining network from LND %v", info.Chains[0].Network)
	}
	return info.IdentityPubkey, chain, network, nil
}

func getInformationFromClnNode(ctx context.Context,
	grpcAddress string,
	certificate []byte,
	key []byte,
	caCertificate []byte) (string, core.Chain, core.Network, error) {

	conn, err := cln_connect.Connect(grpcAddress, certificate, key, caCertificate)
	if err != nil {
		return "", 0, 0, errors.Wrap(err,
			"Can't connect to node to verify public key, check all details including TLS Cert and Macaroon")
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to close gRPC connection.")
		}
	}(conn)

	client := cln.NewNodeClient(conn)
	info, err := client.Getinfo(ctx, &cln.GetinfoRequest{})
	if err != nil {
		return "", 0, 0, errors.Wrap(err, "Obtaining information from LND")
	}

	chain := core.Bitcoin

	var network core.Network
	switch info.Network {
	case "mainnet":
		network = core.MainNet
	case "testnet":
		network = core.TestNet
	case "signet":
		network = core.SigNet
	case "simnet":
		network = core.SimNet
	case "regtest":
		network = core.RegTest
	default:
		return "", 0, 0, errors.Wrapf(err, "Obtaining network from LND %v", info.Network)
	}
	return hex.EncodeToString(info.Id), chain, network, nil
}

func getSignMessagesPermissionFromLndNode(ctx context.Context,
	grpcAddress string,
	tlsCert []byte,
	macaroonFile []byte) (bool, error) {

	conn, err := lnd_connect.Connect(grpcAddress, tlsCert, macaroonFile)
	if err != nil {
		return false, errors.Wrap(err,
			"Can't connect to node to verify public key, check all details including TLS Cert and Macaroon")
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to close gRPC connection.")
		}
	}(conn)

	client := lnrpc.NewLightningClient(conn)
	signMsgReq := lnrpc.SignMessageRequest{
		Msg: []byte("test"),
	}
	signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
	if err != nil {
		return false, nil
	}
	return signMsgResp.Signature != "", nil
}

func processFile(file *multipart.FileHeader) (fileName *string, data []byte, err error) {
	if file != nil {
		fileName = &file.Filename
		caCertificateDataFile, err := file.Open()
		if err != nil {
			return nil, nil, errors.Wrap(err, "opening file")
		}
		data, err = io.ReadAll(caCertificateDataFile)
		if err != nil {
			return nil, nil, errors.Wrap(err, "reading file")
		}
	}
	return fileName, data, nil
}

func getNodeStartDateFromLndNode(ctx context.Context,
	grpcAddress string,
	tlsCert []byte,
	macaroonFile []byte) (*time.Time, error) {

	conn, err := lnd_connect.Connect(grpcAddress, tlsCert, macaroonFile)
	if err != nil {
		return nil, errors.Wrap(err,
			"Can't connect to node to verify public key, check all details including TLS Cert and Macaroon")
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to close gRPC connection.")
		}
	}(conn)

	client := lnrpc.NewLightningClient(conn)
	onChainTxDetails, err := client.GetTransactions(ctx, &lnrpc.GetTransactionsRequest{})

	if err != nil {
		return nil, errors.Wrap(err,
			"Failed to get transactions from node.")
	}

	if len(onChainTxDetails.Transactions) == 0 {
		return nil, nil
	}

	// Chronologically first tx is the last one returned in the array
	firstTx := onChainTxDetails.Transactions[len(onChainTxDetails.Transactions)-1]
	nodeStartTime := time.Unix(firstTx.TimeStamp, 0)

	return &nodeStartTime, nil
}
