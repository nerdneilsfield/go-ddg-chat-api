package ddgchat

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	loggerPkg "github.com/nerdneilsfield/shlogin/pkg/logger"
	"go.uber.org/zap"
)

var logger = loggerPkg.GetLogger()

// 全局变量
var (
	conversations     = make(map[string][]ChatMessage)
	conversationMutex sync.RWMutex
)

func HelloWorld(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}

func RunServer(config *Config) error {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, OPTIONS, DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Use(fiberLogger.New())
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint: "/live",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		ReadinessEndpoint: "/ready",
	}))

	app.Get("/", HelloWorld)

	RegisterRoutes(app, config)

	listenAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	logger.Info("Starting server", zap.String("listen_addr", listenAddr))

	return app.Listen(listenAddr)
}
