package peers

import (
	"database/sql"
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/database"
)

type PeerNode struct {
	NodeId           int                         `json:"nodeId" db:"node_id"`
	Alias            string                      `json:"peerAlias" db:"alias"`
	PublicKey        string                      `json:"pubKey" db:"public_key"`
	TorqNodeId       *int                        `json:"torqNodeId" db:"torq_node_id"`
	TorqNodeAlias    *string                     `json:"torqNodeAlias" db:"torq_node_alias"`
	Setting          *core.NodeConnectionSetting `json:"setting" db:"setting"`
	ConnectionStatus *ConnectionStatus           `json:"connectionStatus" db:"connection_status"`
	Address          *string                     `json:"address" db:"address"`
}

func (p PeerNode) MarshalJSON() ([]byte, error) {
	type Alias PeerNode // create an alias to avoid infinite recursion
	statusStr := ""
	if p.ConnectionStatus != nil {
		statusStr = p.ConnectionStatus.String()
	}
	settingStr := ""
	if p.Setting != nil {
		settingStr = p.Setting.String()
	}
	jsonBytes, err := json.Marshal(&struct {
		*Alias
		ConnectionStatus string `json:"connectionStatus"`
		Setting          string `json:"setting"`
	}{
		Alias:            (*Alias)(&p),
		ConnectionStatus: statusStr,
		Setting:          settingStr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Marshalling PeerNode to JSON.")
	}
	return jsonBytes, nil
}

func GetPeerNodes(db *sqlx.DB, network core.Network) ([]PeerNode, error) {
	var nodes []PeerNode
	err := db.Select(&nodes, `
	SELECT
	n.node_id,
	ne.alias,
	nch.torq_node_id,
	netorq.alias AS torq_node_alias,
	n.public_key,
	nch.connection_status,
	nch.setting
	FROM Node n
	LEFT JOIN (SELECT LAST(node_id, created_on) as node_id, LAST(torq_node_id, created_on) as torq_node_id, LAST(connection_status, created_on) as connection_status, LAST(setting, created_on) as setting, LAST(setting, created_on) as address FROM node_connection_history GROUP BY node_id) nch on nch.node_id = n.node_id
	LEFT JOIN (SELECT LAST(event_node_id, timestamp) as node_id, LAST(alias, timestamp) as alias, LAST(color, timestamp) as color FROM node_event GROUP BY event_node_id) ne ON ne.node_id = n.node_id
	LEFT JOIN (SELECT LAST(event_node_id, timestamp) as node_id, LAST(alias, timestamp) as alias, LAST(color, timestamp) as color FROM node_event GROUP BY event_node_id) netorq ON netorq.node_id = nch.torq_node_id
	WHERE n.network = $1 AND nch.torq_node_id IS NOT NULL;`, network)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []PeerNode{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodes, nil
}
