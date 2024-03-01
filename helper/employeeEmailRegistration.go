// helper/email.go

package helper

import (
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
)

// SendEmployeeAccountNotification sends a welcome email to the employee with a verification link
func SendEmployeeAccountNotificationWithPlainTextPassword(employeeEmail, fullName, username, password string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	sender := smtpUsername
	recipient := employeeEmail
	subject := "Welcome to HR Harmony"
	emailBody := `
    <html>
    <head>
        <style>
            body {
                font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
                background-color: #f9f9f9;
                margin: 0;
                padding: 0;
            }
            .container {
                max-width: 400px;
                margin: 0 auto;
                padding: 20px;
                background-color: #fff;
                box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
                border-radius: 10px;
            }
            h1 {
                text-align: center;
                color: #333;
            }
            .card {
                background-color: #f5f5f5;
                padding: 20px;
                margin-top: 20px;
                border-radius: 5px;
            }
            p {
                font-size: 16px;
                line-height: 1.6;
                margin-top: 10px;
                color: #555;
            }
            strong {
                font-weight: bold;
            }
            a {
                color: #333;
                text-decoration: none;
                font-weight: bold;
            }
            a:hover {
                text-decoration: underline;
            }
            .button {
                display: inline-block;
                padding: 10px 20px;
                font-size: 16px;
                font-weight: bold;
                text-align: center;
                text-decoration: none;
                background-color: #007BFF;
                color: #fff;
                border-radius: 5px;
                display: block;
                margin-top: 20px;
            }
            .button:hover {
                background-color: #0056b3;
            }
            .footer {
                text-align: center;
                margin-top: 20px;
                color: #666;
            }
        </style>
        <!-- Bootstrap CSS -->
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
    </head>
    <body>
        <div class="container">
            <h1>Welcome to HR Harmony for Employee</h1>
            <div class="card">
                <p>Hello, <strong>` + fullName + `</strong>,</p>
                <p>Thank you for choosing HR Harmony as your company's human resource management application. You're now part of our team!</p>
                <p>If you have any questions or need assistance, please don't hesitate to <a href="mailto:support@hrharmony.com">contact our support team</a>.</p>
                <p>Your login credentials:</p>
                <p><strong>Username:</strong> ` + username + `</p>
                <p><strong>Password:</strong> ` + password + `</p>
            </div>
            <a class="button" href="http://localhost:8080/login">Login to HR Harmony</a>
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

// SendEmployeeAccountNotification sends a welcome email to the employee with a verification link
func SendClientAccountNotificationWithPlainTextPassword(employeeEmail, fullName, username, password string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	sender := smtpUsername
	recipient := employeeEmail
	subject := "Welcome to HR Harmony"
	emailBody := `
    <html>
    <head>
        <style>
            body {
                font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
                background-color: #f9f9f9;
                margin: 0;
                padding: 0;
            }
            .container {
                max-width: 400px;
                margin: 0 auto;
                padding: 20px;
                background-color: #fff;
                box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
                border-radius: 10px;
            }
            h1 {
                text-align: center;
                color: #333;
            }
            .card {
                background-color: #f5f5f5;
                padding: 20px;
                margin-top: 20px;
                border-radius: 5px;
            }
            p {
                font-size: 16px;
                line-height: 1.6;
                margin-top: 10px;
                color: #555;
            }
            strong {
                font-weight: bold;
            }
            a {
                color: #333;
                text-decoration: none;
                font-weight: bold;
            }
            a:hover {
                text-decoration: underline;
            }
            .button {
                display: inline-block;
                padding: 10px 20px;
                font-size: 16px;
                font-weight: bold;
                text-align: center;
                text-decoration: none;
                background-color: #007BFF;
                color: #fff;
                border-radius: 5px;
                display: block;
                margin-top: 20px;
            }
            .button:hover {
                background-color: #0056b3;
            }
            .footer {
                text-align: center;
                margin-top: 20px;
                color: #666;
            }
        </style>
        <!-- Bootstrap CSS -->
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
    </head>
    <body>
        <div class="container">
            <h1>Welcome to HR Harmony For Client</h1>
            <div class="card">
                <p>Hello, <strong>` + fullName + `</strong>,</p>
                <p>Thank you for choosing HR Harmony as your company's human resource management application. You're now part of our team!</p>
                <p>If you have any questions or need assistance, please don't hesitate to <a href="mailto:support@hrharmony.com">contact our support team</a>.</p>
                <p>Your login credentials:</p>
                <p><strong>Username:</strong> ` + username + `</p>
                <p><strong>Password:</strong> ` + password + `</p>
            </div>
            <a class="button" href="http://localhost:8080/login">Login to HR Harmony</a>
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
