package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"coral-bot/discord_bot/internal/config"
	"coral-bot/discord_bot/internal/handlers"
	"coral-bot/discord_bot/internal/repository"
	"coral-bot/discord_bot/internal/services"
	"coral-bot/discord_bot/internal/utils"
	"coral-bot/discord_bot/internal/web"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	logger := utils.NewLogger()
	logger.Info("Starting Coral Markets Discord Bot")

	// Initialize repository
	repo := repository.NewInMemorySubscriptionRepository()

	// Initialize services
	marketService := services.NewMarketService(cfg.CoralBackendURL, logger)
	subscriptionService := services.NewSubscriptionService(repo, logger)

	// Initialize command handler
	commandHandler := handlers.NewCommandHandler(marketService, subscriptionService, logger)

	// Initialize webhook handler
	webhookHandler := web.NewWebhookHandler(marketService, subscriptionService, logger)

	// Create Discord session
	dg, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating Discord session: %v", err))
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(commandHandler.HandleInteraction)

	// Set Discord session in webhook handler
	webhookHandler.SetDiscordSession(dg)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		logger.Error(fmt.Sprintf("Error opening connection: %v", err))
		return
	}

	// Register slash commands
	err = commandHandler.RegisterCommands(dg)
	if err != nil {
		logger.Error(fmt.Sprintf("Error registering commands: %v", err))
		return
	}

	// Start webhook server in a goroutine
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	go webhookHandler.StartWebServer(port)

	logger.Info("Coral Markets Discord Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
	logger.Info("Coral Markets Discord Bot stopped")
}
