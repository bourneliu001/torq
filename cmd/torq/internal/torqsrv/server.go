package torqsrv

import (
	"fmt"
	"github.com/lncapital/torq/internal/move_funds"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/lncapital/torq/internal/auth"
	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/internal/categories"
	"github.com/lncapital/torq/internal/channel_history"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/flow"
	"github.com/lncapital/torq/internal/forwards"
	"github.com/lncapital/torq/internal/invoices"
	"github.com/lncapital/torq/internal/lightning"
	"github.com/lncapital/torq/internal/messages"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/on_chain_tx"
	"github.com/lncapital/torq/internal/payments"
	"github.com/lncapital/torq/internal/peers"
	"github.com/lncapital/torq/internal/services"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/internal/views"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/web"
)

func Start(tp *tracesdk.TracerProvider,
	host string,
	port int,
	apiPswd string,
	cookiePath string,
	db *sqlx.DB,
	autoLogin bool,
	prometheusPath string) error {

	var p *Prometheus
	if prometheusPath != "" {
		p = NewPrometheus("gin").SetEnableExemplar(tp != nil)
		p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
			url := c.Request.URL.Path
			for _, p := range c.Params {
				switch p.Key {
				case "channelId":
					url = strings.Replace(url, p.Value, ":channelId", 1)
				case "nodeId":
					url = strings.Replace(url, p.Value, ":nodeId", 1)
				case "network":
					url = strings.Replace(url, p.Value, ":network", 1)
				case "categoryId":
					url = strings.Replace(url, p.Value, ":categoryId", 1)
				case "tagId":
					url = strings.Replace(url, p.Value, ":tagId", 1)
				case "identifier":
					url = strings.Replace(url, p.Value, ":identifier", 1)
				case "chanIds":
					url = strings.Replace(url, p.Value, ":chanIds", 1)
				case "workflowId":
					url = strings.Replace(url, p.Value, ":workflowId", 1)
				case "customSettings":
					url = strings.Replace(url, p.Value, ":customSettings", 1)
				case "statusId":
					url = strings.Replace(url, p.Value, ":statusId", 1)
				case "pingSystems":
					url = strings.Replace(url, p.Value, ":pingSystems", 1)
				case "tableViewId":
					url = strings.Replace(url, p.Value, ":tableViewId", 1)
				}
			}
			return url
		}
		p.SetListenAddress(prometheusPath).SetMetricsPath(nil)
	}

	var ginOtelLogFormatter = func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v traceID=%s\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			trace.SpanContextFromContext(param.Request.Context()).TraceID(),
			param.ErrorMessage,
		)
	}

	r := gin.New()
	switch {
	case tp == nil && p == nil:
		r.Use(otelgin.Middleware("torq-gin"), gin.Logger(), gin.Recovery())
	case tp == nil && p != nil:
		r.Use(otelgin.Middleware("torq-gin"), p.HandlerFunc(), gin.Logger(), gin.Recovery())
	case tp != nil && p == nil:
		r.Use(otelgin.Middleware("torq-gin", otelgin.WithTracerProvider(tp)),
			gin.LoggerWithFormatter(ginOtelLogFormatter), gin.Recovery())
	case tp != nil && p != nil:
		r.Use(otelgin.Middleware("torq-gin", otelgin.WithTracerProvider(tp)), p.HandlerFunc(),
			gin.LoggerWithFormatter(ginOtelLogFormatter), gin.Recovery())
	}

	if err := auth.RefreshCookieFile(cookiePath); err != nil {
		return errors.Wrap(err, "Refreshing cookie file")
	}

	err := auth.CreateSession(r, apiPswd)
	if err != nil {
		return errors.Wrap(err, "Creating Gin Session")
	}

	registerRoutes(r, db, apiPswd, cookiePath, autoLogin)

	fmt.Println("Listening on port " + strconv.Itoa(port))

	if err := r.Run(host + ":" + strconv.Itoa(port)); err != nil {
		return errors.Wrap(err, "Running gin webserver")
	}
	return nil
}

func applyCors(r *gin.Engine) {
	corsConfig := cors.DefaultConfig()
	//hot reload CORS
	corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))
}

// loginKeyGetter is used to force the Login rate
// limiter to limit all requests regardless of IP etc.
func loginKeyGetter(c *gin.Context) string {
	return "login_limiter"
}

// NewLoginRateLimitMiddleware is used to limit login attempts
func NewLoginRateLimitMiddleware() gin.HandlerFunc {
	// Define a limit rate to 10 requests per minute.
	rate, err := limiter.NewRateFromFormatted("10-M")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	store := memory.NewStore()
	return mgin.NewMiddleware(limiter.New(store, rate), mgin.WithKeyGetter(loginKeyGetter))
}

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790.
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}

func registerRoutes(r *gin.Engine, db *sqlx.DB, apiPwd string, cookiePath string, autoLogin bool) {
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	applyCors(r)
	// Websocket
	ws := r.Group("/ws")
	ws.Use(auth.AuthRequired(autoLogin))
	ws.GET("", func(c *gin.Context) {
		err := WebsocketHandler(c, db)
		log.Debug().Msgf("WebsocketHandler: %v", err)
	})

	api := r.Group("/api")

	api.POST("/logout", auth.Logout)

	// Limit login attempts to 10 per minute.
	rl := NewLoginRateLimitMiddleware()
	api.POST("/login", rl, auth.Login(apiPwd))
	api.POST("/cookie-login", rl, auth.CookieLogin(cookiePath))
	api.GET("auto-login-setting", rl, auth.AutoLoginSetting(autoLogin))

	unauthorisedSettingRoutes := api.Group("settings")
	{
		settings.RegisterUnauthenticatedRoutes(unauthorisedSettingRoutes, db)
	}

	unauthorisedServicesRoutes := api.Group("services")
	{
		services.RegisterUnauthenticatedRoutes(unauthorisedServicesRoutes, db)
	}

	api.Use(auth.AuthRequired(autoLogin)).Use(auth.TorqRequired)
	{

		tableViewRoutes := api.Group("/table-views")
		{
			views.RegisterTableViewRoutes(tableViewRoutes, db)
		}

		categoryRoutes := api.Group("/categories")
		{
			categories.RegisterCategoryRoutes(categoryRoutes, db)
		}

		tagRoutes := api.Group("/tags")
		{
			tags.RegisterTagRoutes(tagRoutes, db)
		}

		corridorRoutes := api.Group("/corridors")
		{
			corridors.RegisterCorridorRoutes(corridorRoutes, db)
		}

		paymentRoutes := api.Group("/payments")
		{
			payments.RegisterPaymentsRoutes(paymentRoutes, db)
		}

		invoiceRoutes := api.Group("/invoices")
		{
			invoices.RegisterInvoicesRoutes(invoiceRoutes, db)
		}

		onChainTx := api.Group("/on-chain-tx")
		{
			on_chain_tx.RegisterOnChainTxsRoutes(onChainTx, db)
		}

		peerRoutes := api.Group("/peers")
		{
			peers.RegisterPeerRoutes(peerRoutes, db)
		}

		nodeRoutes := api.Group("/nodes")
		{
			nodes.RegisterNodeRoutes(nodeRoutes, db)
		}

		channelRoutes := api.Group("/channels")
		{
			channel_history.RegisterChannelHistoryRoutes(channelRoutes, db)
			channels.RegisterChannelRoutes(channelRoutes, db)
		}

		forwardRoutes := api.Group("/forwards")
		{
			forwards.RegisterForwardsRoutes(forwardRoutes, db)
		}

		flowRoutes := api.Group("/flow")
		{
			flow.RegisterFlowRoutes(flowRoutes, db)
		}

		lightningRoutes := api.Group("/lightning")
		{
			lightning.RegisterLightningRoutes(lightningRoutes, db)
		}

		workflowRoutes := api.Group("/workflows")
		{
			workflows.RegisterWorkflowRoutes(workflowRoutes, db)
		}

		automationRoutes := api.Group("/automation")
		{
			automation.RegisterAutomationRoutes(automationRoutes, db)
		}

		messageRoutes := api.Group("messages")
		{
			messages.RegisterMessagesRoutes(messageRoutes)
		}

		settingRoutes := api.Group("settings")
		{
			settings.RegisterSettingRoutes(settingRoutes, db)
		}

		moveFundsRoutes := api.Group("move-funds")
		{
			move_funds.RegisterMoveFundsRoutes(moveFundsRoutes)
		}

		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	registerStaticRoutes(r)
}

func registerStaticRoutes(r *gin.Engine) {
	embeddedFS := web.NewStaticFileSystem()
	r.NoRoute(func(c *gin.Context) {
		log.Warn().Msg("No route")
		path := c.Request.URL.Path

		knownAssetList := []string{
			"/favicon.ico",
			"/favicon-16x16.png",
			"/favicon-32x32.png",
			"/mstile-150x150.png",
			"/safari-pinned-tab.svg",
			"/android-chrome-192x192.png",
			"/android-chrome-512x512.png",
			"/apple-touch-icon.png",
			"/browserconfig.xml",
			"/manifest.json",
			"/robots.txt",
		}

		for _, knownAsset := range knownAssetList {
			if strings.HasSuffix(path, knownAsset) {
				c.FileFromFS(knownAsset, embeddedFS)
				return
			}
		}

		// probably a file, this might not be bulletproof
		if strings.Contains(path, "/static/") && strings.Contains(path, ".") &&
			(strings.Contains(path, "css") || strings.Contains(path, "js") || strings.Contains(path, "media")) {
			parts := strings.Split(path, "/")
			c.FileFromFS("static/"+parts[len(parts)-2]+"/"+parts[len(parts)-1], embeddedFS)
			return
		}

		// locales json files
		if strings.Contains(path, "/locales/") && strings.Contains(path, ".json") {
			parts := strings.Split(path, "/")
			c.FileFromFS("locales/"+parts[len(parts)-1], embeddedFS)
			return
		}

		// the reason that we can't use c.FileFromFS for index.html is because golang will return a redirect instead of serving the file
		// so we have to manually read the file and send it out
		// https://stackoverflow.com/questions/43527073/golang-static-stop-index-html-redirection
		f, err := embeddedFS.Open("index.html")
		if err != nil {
			log.Panic().Err(err).Msg("Couldn't read index.html")
			panic(err)
		}
		http.ServeContent(c.Writer, c.Request, "index.html", time.Now(), f)
	})
}
