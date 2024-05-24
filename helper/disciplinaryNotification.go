package helper

import (
	"fmt"
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
)

// SendDisciplinaryNotification mengirimkan email notifikasi kepada karyawan saat data disciplinary dibuat
func SendDisciplinaryNotification(employeeEmail, employeeFullName, caseName, subject, caseDate, description string) error {
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
			<h1>Notifikasi Disciplinary</h1>
			<p>Halo %s,</p>
			<p>Anda telah menerima disciplinary baru dengan rincian sebagai berikut:</p>
			<p>Nama Kasus: <strong>%s</strong></p>
			<p>Subjek: <strong>%s</strong></p>
			<p>Tanggal Kasus: <strong>%s</strong></p>
			<p>Deskripsi: <strong>%s</strong></p>
			<div class="footer">
				<p>&copy; 2024 HR Harmony. All rights reserved.</p>
			</div>
		</div>
	</body>
	</html>
	`, employeeFullName, caseName, subject, caseDate, description)

	// Set konfigurasi email
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	sender := smtpUsername
	recipient := employeeEmail
	subjectEmail := "Notifikasi Disciplinary"

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
