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
	cfg := config.LoadConfig()

	logger := utils.NewLogger()
	logger.Info("Starting Coral Markets Discord Bot")

	repo := repository.NewInMemorySubscriptionRepository()

	marketService := services.NewMarketService(cfg.CoralBackendURL, logger)
	subscriptionService := services.NewSubscriptionService(repo, logger)

	commandHandler := handlers.NewCommandHandler(marketService, subscriptionService, logger)

	webhookHandler := web.NewWebhookHandler(marketService, subscriptionService, logger)

	dg, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating Discord session: %v", err))
		return
	}

	dg.AddHandler(commandHandler.HandleInteraction)

	webhookHandler.SetDiscordSession(dg)

	err = dg.Open()
	if err != nil {
		logger.Error(fmt.Sprintf("Error opening connection: %v", err))
		return
	}

	err = commandHandler.RegisterCommands(dg)
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
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
	logger.Info("Coral Markets Discord Bot stopped")
}
