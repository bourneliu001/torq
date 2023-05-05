package move_funds

import (
	"github.com/gin-gonic/gin"
)

func RegisterMoveFundsRoutes(r *gin.RouterGroup) {
	r.POST("off-chain", func(c *gin.Context) { moveFundsOffChainHandler(c) })
}
