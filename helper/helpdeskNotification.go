package helper

import (
	"fmt"
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
)

// SendHelpdeskNotification mengirimkan email notifikasi kepada karyawan setelah admin membuat tiket helpdesk untuk mereka
func SendHelpdeskNotification(employeeEmail, fullName, subject, description string) error {
	// Konstruksi isi email
	emailBody := fmt.Sprintf(`
	<html>
	<head>
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f4f4f4;
				margin: 0;
				padding: 20px;
			}
			.container {
				background-color: #fff;
				padding: 30px;
				border-radius: 5px;
				box-shadow: 0 2px 5px rgba(0,0,0,0.1);
			}
			h1 {
				color: #333;
			}
			p {
				font-size: 16px;
				line-height: 1.6;
				margin: 10px 0;
			}
			strong {
				font-weight: bold;
			}
			.footer {
				text-align: center;
				margin-top: 20px;
				color: #666;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Notifikasi Tiket Helpdesk Baru</h1>
			<p>Halo %s,</p>
			<p>Anda memiliki tiket helpdesk baru dengan rincian sebagai berikut:</p>
			<p>Subjek: <strong>%s</strong></p>
			<p>Deskripsi: <strong>%s</strong></p>
			<p>Status: <strong>Open</strong></p>
			<p>Silakan hubungi bagian terkait jika Anda memiliki pertanyaan lebih lanjut.</p>
			<div class="footer">
				<p>&copy; 2024 HR Harmony. All rights reserved.</p>
			</div>
		</div>
	</body>
	</html>
	`, fullName, subject, description)

	// Set konfigurasi email
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	sender := smtpUsername
	recipient := employeeEmail
	subjectEmail := "Notifikasi Tiket Helpdesk Baru"

	// Buat pesan email
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subjectEmail)
	m.SetBody("text/html", emailBody)

	// Konfigurasi dialer
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return err
	}
	d := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)

	// Kirim email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

// SendHelpdeskNotification mengirimkan email notifikasi kepada karyawan setelah admin membuat tiket helpdesk untuk mereka
func SendHelpdeskNotificationStatus(employeeEmail, fullName, subject, description, status string) error {
	// Konstruksi isi email
	emailBody := fmt.Sprintf(`
	<html>
	<head>
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f4f4f4;
				margin: 0;
				padding: 20px;
			}
			.container {
				background-color: #fff;
				padding: 30px;
				border-radius: 5px;
				box-shadow: 0 2px 5px rgba(0,0,0,0.1);
			}
			h1 {
				color: #333;
			}
			p {
				font-size: 16px;
				line-height: 1.6;
				margin: 10px 0;
			}
			strong {
				font-weight: bold;
			}
			.footer {
				text-align: center;
				margin-top: 20px;
				color: #666;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Notifikasi Tiket Helpdesk Baru</h1>
			<p>Halo %s,</p>
			<p>Anda memiliki tiket helpdesk baru dengan rincian sebagai berikut:</p>
			<p>Subjek: <strong>%s</strong></p>
			<p>Deskripsi: <strong>%s</strong></p>
			<p>Status: <strong>%s</strong></p>
			<p>Silakan hubungi bagian terkait jika Anda memiliki pertanyaan lebih lanjut.</p>
			<div class="footer">
				<p>&copy; 2024 HR Harmony. All rights reserved.</p>
			</div>
		</div>
	</body>
	</html>
	`, fullName, subject, description, status)

	// Set konfigurasi email
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	sender := smtpUsername
	recipient := employeeEmail
	subjectEmail := "Notifikasi Tiket Helpdesk Baru"

	// Buat pesan email
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subjectEmail)
	m.SetBody("text/html", emailBody)

	// Konfigurasi dialer
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return err
	}
	d := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)

	// Kirim email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
