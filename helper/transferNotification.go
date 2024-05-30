package helper

import (
	"fmt"
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
	"strings"
)

// FormatToIDR mengonversi jumlah dalam format float menjadi format mata uang Rupiah yang sesuai dengan standar Indonesia
func FormatToIDR(amount float64) string {
	// Format angka ke string dengan 2 digit desimal
	amountStr := strconv.FormatFloat(amount, 'f', 2, 64)

	// Pisahkan angka berdasarkan titik desimal
	parts := strings.Split(amountStr, ".")

	// Format bagian integer dengan pemisah ribuan
	integerPart := addThousandSeparator(parts[0])

	// Gabungkan bagian integer yang diformat dengan bagian desimal (jika ada)
	formattedAmount := "Rp " + integerPart
	if len(parts) > 1 {
		formattedAmount += "," + parts[1]
	}

	return formattedAmount
}

// AddThousandSeparator menambahkan pemisah ribuan ke string angka
func addThousandSeparator(number string) string {
	var result string
	length := len(number)
	for i, c := range number {
		result += string(c)
		if (length-i-1)%3 == 0 && i != length-1 {
			result += "."
		}
	}
	return result
}

// SendSalaryTransferNotification mengirimkan email notifikasi kepada karyawan tentang transfer gaji yang berhasil
func SendSalaryTransferNotification(employeeEmail, fullName string, finalSalary float64) error {
	// Format jumlah gaji ke mata uang Rupiah
	formattedBasicSalary := FormatToIDR(finalSalary)

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
			<h1>Salary Transfer Notification</h1>
			<p>Halo, <strong>%s</strong>,</p>
			<p>We are pleased to inform you that your salary for this month has been successfully transferred.</p>
			<p>The amount transferred is: <strong>%s</strong></p>
			<p>If you have any questions or concerns, please contact the HR department.</p>
			<div class="footer">
				<p>&copy; 2024 HR Harmony. All rights reserved.</p>
			</div>
		</div>
	</body>
	</html>
	`, fullName, formattedBasicSalary)

	// Set konfigurasi email
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	sender := smtpUsername
	recipient := employeeEmail
	subject := "HR Harmony: Salary Transfer Notification"

	// Buat pesan email
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
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
