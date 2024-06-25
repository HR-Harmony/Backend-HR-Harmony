package helper

import (
	"bytes"
	"fmt"
	"github.com/go-gomail/gomail"
	"github.com/jung-kurt/gofpdf"
	"io"
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

func generateSalarySlipPDF(fullName string, finalSalary, lateDeduction, earlyLeavingDeduction, overtimePay, totalLoanDeduction float64) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Salary Slip")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Employee Name: %s", fullName))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Final Salary: %s", FormatToIDR(finalSalary)))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Late Deduction: %s", FormatToIDR(lateDeduction)))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Early Leaving Deduction: %s", FormatToIDR(earlyLeavingDeduction)))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Overtime Pay: %s", FormatToIDR(overtimePay)))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Loan Deduction: %s", FormatToIDR(totalLoanDeduction)))
	pdf.Ln(10)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func SendSalaryTransferNotification(employeeEmail, fullName string, finalSalary, lateDeduction, earlyLeavingDeduction, overtimePay, totalLoanDeduction float64) error {
	// Format jumlah gaji ke mata uang Rupiah
	formattedFinalSalary := FormatToIDR(finalSalary)
	formattedLateDeduction := FormatToIDR(lateDeduction)
	formattedEarlyLeavingDeduction := FormatToIDR(earlyLeavingDeduction)
	formattedOvertimePay := FormatToIDR(overtimePay)
	formattedLoanDeduction := FormatToIDR(totalLoanDeduction)

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
			<p>Kami ingin memberitahukan bahwa gaji Anda untuk bulan ini telah berhasil ditransfer.</p>
			<p>Rincian gaji Anda adalah sebagai berikut:</p>
			<p>Gaji Dasar: <strong>%s</strong></p>
			<p>Potongan Keterlambatan: <strong>%s</strong></p>
			<p>Potongan Early Leaving: <strong>%s</strong></p>
			<p>Tambahan Overtime: <strong>%s</strong></p>
			<p>Potongan Pinjaman: <strong>%s</strong></p>
			<p>Gaji Akhir yang Ditranfer: <strong>%s</strong></p>
			<p>Jika Anda memiliki pertanyaan atau kekhawatiran, silakan hubungi departemen HR.</p>
			<div class="footer">
				<p>&copy; 2024 HR Harmony. All rights reserved.</p>
			</div>
		</div>
	</body>
	</html>
	`, fullName, formattedFinalSalary, formattedLateDeduction, formattedEarlyLeavingDeduction, formattedOvertimePay, formattedLoanDeduction, formattedFinalSalary)

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

	// Generate slip gaji dalam bentuk PDF
	pdfBytes, err := generateSalarySlipPDF(fullName, finalSalary, lateDeduction, earlyLeavingDeduction, overtimePay, totalLoanDeduction)
	if err != nil {
		return err
	}

	// Lampirkan slip gaji PDF ke email
	m.Attach("SalarySlip.pdf", gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(pdfBytes)
		return err
	}))

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

/*
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
*/

// SendLoanApprovalNotification mengirimkan email notifikasi kepada karyawan tentang persetujuan pinjaman yang berhasil
func SendLoanApprovalNotification(employeeEmail, fullName string, amount float64) error {
	// Format jumlah pinjaman ke mata uang Rupiah
	formattedAmount := FormatToIDR(amount)

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
			<h1>Loan Approval Notification</h1>
			<p>Halo, <strong>%s</strong>,</p>
			<p>We are pleased to inform you that your loan request has been approved and the amount has been successfully transferred to your account.</p>
			<p>The amount transferred is: <strong>%s</strong></p>
			<p>If you have any questions or concerns, please contact the HR department.</p>
			<div class="footer">
				<p>&copy; 2024 HR Harmony. All rights reserved.</p>
			</div>
		</div>
	</body>
	</html>
	`, fullName, formattedAmount)

	// Set konfigurasi email
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	sender := smtpUsername
	recipient := employeeEmail
	subject := "HR Harmony: Loan Approval Notification"

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
