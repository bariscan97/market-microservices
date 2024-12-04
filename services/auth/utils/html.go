package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
)

func ResetPasswordTemplate(token string) string {
	
	htmlContent, err := os.ReadFile("../templates/resetpassword.html")
	if err != nil {
		log.Fatalf("Failed to read HTML file: %s", err)
	}

	forgotPasswordLink := fmt.Sprintf("http://localhost:%s/api/auth/resetpassword?token=%s" ,"8080" , token)

	str := string(htmlContent)

	str = strings.ReplaceAll(str,"{{{ForgotPasswordLink}}}", forgotPasswordLink)

	return str
}


func RegisterVerifyTempalte(token string) bytes.Buffer {
	
	htmlContent, err := os.ReadFile("../templates/verify_template.html")
	
	if err != nil {
		log.Fatalf("Failed to read HTML file: %s", err)
	}

	verificationLink := fmt.Sprintf("http://localhost:%s/api/auth/verify?token=%s" ,"8080", token)

	tmpl, err := template.New("email").Parse(string(htmlContent))
	if err != nil {
		log.Fatalf("Failed to parse template: %s", err)
	}

	var body bytes.Buffer

	err = tmpl.Execute(&body, struct {
        VerificationLink string
    }{
        VerificationLink: verificationLink,
    })
	
	if err != nil {
		log.Fatalf("Failed to execute template: %s", err)
	}

	return body
}

func ForgotPasswordTempalte(token string) bytes.Buffer {
	
	htmlContent, err := os.ReadFile("../templates/forgotpassword.html")
	if err != nil {
		log.Fatalf("Failed to read HTML file: %s", err)
	}



	forgotPasswordLink := fmt.Sprintf("http://localhost:%s/api/auth/resetpassword?token=%s", "8080", token)

	tmpl, err := template.New("email").Parse(string(htmlContent))
	if err != nil {
		log.Fatalf("Failed to parse template: %s", err)
	}

	var body bytes.Buffer

	err = tmpl.Execute(&body, struct {
        ForgotPasswordLink string
    }{
        ForgotPasswordLink: forgotPasswordLink,
    })
	
	if err != nil {
		log.Fatalf("Failed to execute template: %s", err)
	}
	
	return body
}


