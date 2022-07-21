package invoices

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	qp "github.com/lncapital/torq/internal/query_parser"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	"strconv"
)

func getInvoicesHandler(c *gin.Context, db *sqlx.DB) {

	// Filter parser with whitelisted columns
	var filter sq.Sqlizer
	filterParam := c.Query("filter")
	var err error
	if filterParam != "" {
		filter, err = qp.ParseFilterParam(filterParam, []string{
			"creation_date",
			"settle_date",
			"add_index",
			"settle_index",
			"payment_request",
			"destination_pub_key",
			"r_hash",
			"r_preimage",
			"memo",
			"value_msat",
			"amt_paid_msat",
			"invoice_state",
			"is_rebalance",
			"is_keysend",
			"is_amp",
			"payment_addr",
			"fallback_addr",
			"updated_on",
			"expiry",
			"cltv_expiry",
			"private",
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	var sort []string
	sortParam := c.Query("order")
	if sortParam != "" {
		// Order parser with whitelisted columns
		sort, err = qp.ParseOrderParams(
			sortParam,
			[]string{
				"creation_date",
				"settle_date",
				"add_index",
				"settle_index",
				"memo",
				"value_msat",
				"amt_paid_msat",
				"invoice_state",
				"is_keysend",
				"is_amp",
				"updated_on",
				"expiry",
				"private",
			})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	var limit uint64 = 100
	if c.Query("limit") != "" {
		limit, err = strconv.ParseUint(c.Query("limit"), 10, 64)
		switch err.(type) {
		case nil:
			break
		case *strconv.NumError:
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Limit must be a positive number"})
			return
		default:
			server_errors.LogAndSendServerError(c, err)
		}
		if limit == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Limit must be a at least 1"})
			return
		}
	}

	var offset uint64 = 0
	if c.Query("offset") != "" {
		offset, err = strconv.ParseUint(c.Query("offset"), 10, 64)
		switch err.(type) {
		case nil:
			break
		case *strconv.NumError:
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Offset must be a positive number"})
			return
		default:
			server_errors.LogAndSendServerError(c, err)
		}
	}

	r, err := getInvoices(db, filter, sort, limit, offset)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}

func getInvoiceHandler(c *gin.Context, db *sqlx.DB) {

	r, err := getInvoiceDetails(db, c.Param("identifier"))
	switch err.(type) {
	case nil:
		break
	case ErrInvoiceNotFound:
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error(), "Identifier": c.Param("identifier")})
		return
	default:
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}

func RegisterInvoicesRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getInvoicesHandler(c, db) })
	r.GET(":identifier", func(c *gin.Context) { getInvoiceHandler(c, db) })
}
