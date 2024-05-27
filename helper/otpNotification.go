package helper

import (
	"os"
	"strconv"
	"time"

	"github.com/go-gomail/gomail"
)

func SendPasswordResetOTP(userEmail string, fullname string, otp string, expiredAt time.Time) error {
	// Konversi expiredAt ke Waktu Indonesia Barat (WIB)
	wibExpiredAt := convertToWIB(expiredAt).Format("02 Jan 2006 15:04:05 MST")

	// Mengambil nilai dari environment variables
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	// Konfigurasi pengiriman email
	sender := smtpUsername
	recipient := userEmail
	subject := "Password Reset OTP Notification"
	emailBody := `
	<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset OTP</title>
    <style>
        body {
            font-family: 'Arial', sans-serif;
            background: linear-gradient(180deg, #007BFF, #00BFFF);
            color: #fff;
            margin: 0;
            padding: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
        }
        .container {
            max-width: 600px;
            width: 100%;
            background-color: #fff;
            box-shadow: 0 0 20px rgba(0, 0, 0, 0.2);
            border-radius: 10px;
            overflow: hidden;
            text-align: center;
            margin: 0 auto; /* Menempatkan container di tengah */
        }
        .header {
            background-color: #007BFF;
            color: #fff;
            padding: 20px;
            border-bottom: 1px solid #ddd;
        }
        h1 {
            margin: 0;
            color: #333;
            font-size: 28px;
        }
        .logo {
            text-align: center;
            margin-top: 20px;
        }
        .logo img {
            width: 120px;
            height: 120px;
            border-radius: 50%;
            border: 3px solid #007BFF;
            transition: transform 0.3s ease-in-out;
        }
        .logo img:hover {
            transform: scale(1.1);
        }
        .message {
            padding: 20px;
        }
        p {
            font-size: 18px;
            margin-top: 15px;
            color: #555;
            line-height: 1.5;
        }
        .footer {
            text-align: center;
            padding: 20px;
            color: #666;
            font-size: 14px;
            border-top: 1px solid #ddd;
        }
        a {
            text-decoration: none;
            color: #007BFF;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset OTP</h1>
        </div>
        <div class="logo">
            <img src="https://i.ibb.co/k1nSQZY/HR-Hamony.png" alt="HR Harmony Logo">
        </div>
        <div class="message">
            <p>Hello, <strong>` + fullname + `</strong>,</p>
            <p>You have requested to reset your password. Please use the OTP below to proceed:</p>
            <h2 style="font-size: 36px; color: #007BFF;">` + otp + `</h2>
            <p>This OTP will expire at ` + wibExpiredAt + `.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 HR Harmony. All rights reserved. | <a href="https://hr-harmony.seculab.space" target="_blank">HR Harmony</a></p>
        </div>
    </div>
</body>
</html>
	`

	// Convert the SMTP port from string to integer
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	// Set pesan dalam format HTML
	m.SetBody("text/html", emailBody)

	d := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func SendPasswordChangedNotification(userEmail string, fullname string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	sender := smtpUsername
	recipient := userEmail
	subject := "Password Changed Notification"
	emailBody := `
	<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Changed Notification</title>
    <style>
        body {
            font-family: 'Arial', sans-serif;
            background: linear-gradient(180deg, #007BFF, #00BFFF);
            color: #fff;
            margin: 0;
            padding: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
        }
        .container {
            max-width: 600px;
            width: 100%;
            background-color: #fff;
            box-shadow: 0 0 20px rgba(0, 0, 0, 0.2);
            border-radius: 10px;
            overflow: hidden;
            text-align: center;
            margin: 0 auto;
        }
        .header {
            background-color: #007BFF;
            color: #fff;
            padding: 20px;
            border-bottom: 1px solid #ddd;
        }
        h1 {
            margin: 0;
            color: #333;
            font-size: 28px;
        }
        .logo {
            text-align: center;
            margin-top: 20px;
        }
        .logo img {
            width: 120px;
            height: 120px;
            border-radius: 50%;
            border: 3px solid #007BFF;
            transition: transform 0.3s ease-in-out;
        }
        .logo img:hover {
            transform: scale(1.1);
        }
        .message {
            padding: 20px;
        }
        p {
            font-size: 18px;
            margin-top: 15px;
            color: #555;
            line-height: 1.5;
        }
        .footer {
            text-align: center;
            padding: 20px;
            color: #666;
            font-size: 14px;
            border-top: 1px solid #ddd;
        }
        a {
            text-decoration: none;
            color: #007BFF;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Changed Notification</h1>
        </div>
        <div class="logo">
            <img src="https://i.ibb.co/k1nSQZY/HR-Hamony.png" alt="HR Harmony Logo">
        </div>
        <div class="message">
            <p>Hello, <strong>` + fullname + `</strong>,</p>
            <p>Your password has been changed successfully. If you did not perform this action, please contact support immediately.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 HR Harmony. All rights reserved. | <a href="https://hr-harmony.seculab.space" target="_blank">HR Harmony</a></p>
        </div>
    </div>
</body>
</html>
	`

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", emailBody)

	d := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
