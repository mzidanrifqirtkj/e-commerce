package cmd

import (
	"encoding/json"
	"log"
	"net/smtp"
	"strings"

	"user-service/config"

	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "start email worker consuming rabbitmq queues",
	Long:  "consume email queues (email_verification, forgot-password) and send via SMTP",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()

		conn, err := cfg.NewRabbitMQ()
		if err != nil {
			log.Fatalf("[Worker-1] Failed to connect RabbitMQ: %v", err)
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			log.Fatalf("[Worker-2] Failed to open channel: %v", err)
		}
		defer ch.Close()

		for _, queue := range []string{"email_verification", "forgot-password"} {
			msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
			if err != nil {
				log.Fatalf("[Worker-3] Failed to consume %s: %v", queue, err)
			}
			go handleMessages(cfg, msgs, queue)
		}

		log.Println("[Worker-4] Email worker started. Waiting for messages...")
		select {}
	},
}

func handleMessages(cfg *config.Config, msgs <-chan amqp.Delivery, queue string) {
	for d := range msgs {
		var notif struct {
			Email   string `json:"email"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(d.Body, &notif); err != nil {
			log.Printf("[Worker-5] Failed to unmarshal from %s: %v", queue, err)
			continue
		}

		subject := "Email Verification"
		if queue == "forgot-password" {
			subject = "Forgot Password"
		}

		msg := []byte("To: " + notif.Email + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" + notif.Message + "\r\n")

		addr := cfg.SMTP.Host + ":" + cfg.SMTP.Port
		auth := smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Password, cfg.SMTP.Host)

		// ponytail: net/smtp plain only, no TLS/HTML; upgrade when real delivery needed
		if err := smtp.SendMail(addr, auth, cfg.SMTP.From, []string{notif.Email}, msg); err != nil {
			log.Printf("[Worker-6] Failed to send email to %s: %v", notif.Email, err)
			continue
		}
		log.Printf("[Worker-7] Email sent to %s via %s", notif.Email, queue)
	}
}

func init() {
	rootCmd.AddCommand(workerCmd)
}

var _ = strings.TrimSpace