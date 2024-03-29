package helper

import (
	"fmt"
	"github.com/go-gomail/gomail"
	"hrsale/models"
	"os"
	"strconv"
)

func GetCooperationEmailBody(message models.CooperationMessage) string {
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>New Cooperation Message</title>
			<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
			<style>
				body {
					background-color: #f8f9fa;
					color: #495057;
					font-family: 'Arial', sans-serif;
					padding: 20px;
				}
				.card {
					background-color: #fff;
					border: 1px solid rgba(0, 0, 0, 0.125);
					border-radius: 0.25rem;
					box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
					padding: 20px;
					margin-top: 20px;
				}
				h2 {
					color: #007bff;
				}
				strong {
					color: #007bff;
				}
				.logo {
					display: block;
					margin: 0 auto;
					width: 100px; /* Ubah ukuran gambar sesuai kebutuhan */
					border-radius: 200px;
				}
			</style>
		</head>
		<body>
			<div class="card">
				<a href="https://imgbb.com/"><img src="https://i.ibb.co/k1nSQZY/HR-Hamony.png" alt="HR-Hamony" border="0"></a>
				<h2 class="mt-4">New Cooperation Message</h2>
				<p><strong>Name:</strong> %s %s</p>
				<p><strong>Email:</strong> %s</p>
				<p><strong>Phone Number:</strong> %s</p>
				<p><strong>Message:</strong> %s</p>
			</div>
		</body>
		</html>
	`,
		message.FirstName, message.LastName, message.Email, message.PhoneNumber, message.Message)

	return body
}

func GetUserCooperationEmailBody(message models.CooperationMessage) string {
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Your Cooperation Message</title>
			<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
			<style>
				body {
					background-color: #f8f9fa;
					color: #495057;
					font-family: 'Arial', sans-serif;
					padding: 20px;
				}
				.card {
					background-color: #fff;
					border: 1px solid rgba(0, 0, 0, 0.125);
					border-radius: 0.25rem;
					box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
					padding: 20px;
					margin-top: 20px;
				}
				h2 {
					color: #007bff;
				}
				strong {
					color: #007bff;
				}
				.logo {
					display: block;
					margin: 0 auto;
					width: 100px; /* Ubah ukuran gambar sesuai kebutuhan */
					border-radius: 200px;
				}
			</style>
		</head>
		<body>
			<div class="card">
				<a href="https://imgbb.com/"><img src="https://i.ibb.co/k1nSQZY/HR-Hamony.png" alt="HR-Hamony" border="0"></a>
				<h2 class="mt-4">Thank you for your Message</h2>
				<p>We will get back to you as soon as possible.</p>

				<h3 class="mt-4">Your Message:</h3>
				<p><strong>Name:</strong> %s %s</p>
				<p><strong>Email:</strong> %s</p>
				<p><strong>Phone Number:</strong> %s</p>
				<p><strong>Message:</strong> %s</p>
			</div>
		</body>
		</html>
	`,
		message.FirstName, message.LastName, message.Email, message.PhoneNumber, message.Message)

	return body
}

func SendEmailToUser(email string, subject string, body string) error {
	// Baca variabel lingkungan untuk konfigurasi SMTP
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	// Konversi smtpPortStr menjadi int
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUsername)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
