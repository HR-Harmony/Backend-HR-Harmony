package helper

import (
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
)

func SendWelcomeEmail(adminEmail, fullName, verificationToken string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	sender := smtpUsername
	recipient := adminEmail
	subject := "Welcome to HR Harmony"
	verificationLink := "https://backend-hr-harmony.seculab.space/verify?token=" + verificationToken
	emailBody := `
    <html>
    <head>
        <link href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">
        <style>
            body {
                font-family: 'Arial', sans-serif;
                background-color: #f5f5f5;
                margin: 0;
                padding: 0;
            }
            .container {
                max-width: 600px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
                border-radius: 5px;
            }
            h1 {
                text-align: center;
                color: #333;
            }
            .message {
                background-color: #f9f9f9;
                padding: 15px;
                border: 1px solid #ddd;
                border-radius: 5px;
            }
            p {
                font-size: 16px;
                margin-top: 10px;
                line-height: 1.6;
            }
            strong {
                font-weight: bold;
            }
            .footer {
                text-align: center;
                margin-top: 20px;
                color: #666;
            }
            .btn-verify-email {
                background-color: #1E90FF;
                color: #fff;
                padding: 10px 20px;
                border-radius: 5px;
                text-decoration: none;
                display: inline-block;
                margin: 20px auto;
            }
            .btn-verify-email:hover {
                background-color: #007BFF;
            }
            .logo {
                text-align: center;
                margin-top: 20px;
            }
            .logo img {
                width: 120px;
                height: 120px;
                border-radius: 50%;
                border: 3px solid #1E90FF;
                transition: transform 0.3s ease-in-out;
                margin: 0 auto;
                display: block;
            }
            .logo img:hover {
                transform: scale(1.1);
            }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="logo">
                <a href="https://imgbb.com/"><img src="https://i.ibb.co/k1nSQZY/HR-Hamony.png" alt="HR-Hamony" border="0"></a>
            </div>
            <h1>Welcome to HR Harmony</h1>
            <div class="message">
                <p>Hello, <strong>` + fullName + `</strong>,</p>
                <p>Thank you for choosing HR Harmony as your company's human resource management application. You're now part of our team!</p>
                <p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>
                <p><strong>Support Team:</strong> <a href="mailto:hriscloud@gmail.com">hriscloud@gmail.com</a></p>
                <a href="` + verificationLink + `" class="btn btn-verify-email">Verify Email</a>
            </div>
            <div class="footer">
                <p>&copy; 2023 HR Harmony. All rights reserved.</p>
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
