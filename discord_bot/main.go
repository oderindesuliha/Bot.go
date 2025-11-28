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
    appConfig := config.LoadConfig()

	logger := utils.NewLogger()
	logger.Info("Starting Coral Markets Discord Bot")

    subscriptionRepo := repository.NewInMemorySubscriptionRepository()

    marketService := services.NewMarketService(appConfig.CoralBackendURL, logger)
    subscriptionService := services.NewSubscriptionService(subscriptionRepo, logger)

	commandHandler := handlers.NewCommandHandler(marketService, subscriptionService, logger)

	webhookHandler := web.NewWebhookHandler(marketService, subscriptionService, logger)

    discordSession, err := discordgo.New("Bot " + appConfig.DiscordBotToken)
    if err != nil {
        logger.Error(fmt.Sprintf("Error creating Discord session: %v", err))
        return
    }

    discordSession.AddHandler(commandHandler.HandleInteraction)

    webhookHandler.SetDiscordSession(discordSession)

    err = discordSession.Open()
    if err != nil {
        logger.Error(fmt.Sprintf("Error opening connection: %v", err))
        return
    }

    err = commandHandler.RegisterCommands(discordSession)
    if err != nil {
        logger.Error(fmt.Sprintf("Error registering commands: %v", err))
        return
    }

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	go webhookHandler.StartWebServer(port)

	logger.Info("Coral Markets Discord Bot is now running. Press CTRL-C to exit.")
    shutdownSignal := make(chan os.Signal, 1)
    signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-shutdownSignal

    discordSession.Close()
    logger.Info("Coral Markets Discord Bot stopped")
}
