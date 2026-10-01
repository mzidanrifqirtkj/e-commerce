package cmd

import (
	"encoding/json"
	"log"
	"net/smtp"

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
			// autoAck is false so a message is only acknowledged after the email
			// is actually sent; a failure is requeued instead of being lost.
			msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
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
			// Unparseable payload can never succeed on retry, so drop it
			// instead of requeueing forever.
			log.Printf("[Worker-5] Failed to unmarshal from %s: %v", queue, err)
			_ = d.Nack(false, false)
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
		// ponytail: a send failure requeues with no backoff and no attempt cap;
		// add a retry limit or a dead-letter queue before real traffic.
		if err := smtp.SendMail(addr, auth, cfg.SMTP.From, []string{notif.Email}, msg); err != nil {
			log.Printf("[Worker-6] Failed to send email to %s: %v", notif.Email, err)
			_ = d.Nack(false, true)
			continue
		}

		log.Printf("[Worker-7] Email sent to %s via %s", notif.Email, queue)
		if err := d.Ack(false); err != nil {
			log.Printf("[Worker-8] Failed to ack %s: %v", queue, err)
		}
	}
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
