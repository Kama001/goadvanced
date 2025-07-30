// Sensitive Data Redactor
// ---------------------------
// Input: File containing log lines with emails, phone numbers, credit card numbers
// Task: Mask sensitive fields while preserving format. Output sanitized version.

// 2025-07-30 09:10:55 - Login attempt by a***@g***.com
// 2025-07-30 09:12:20 - OTP sent to mobile number: +91-**********
// 2025-07-30 09:15:03 - Credit card used: ****-****-****-7356
// 2025-07-30 09:17:45 - New user signed up: b***@p***********.com
// 2025-07-30 09:20:10 - Emergency contact: (***) ***-****
// 2025-07-30 09:21:12 - Transaction failed for card ****-****-****-0004
// 2025-07-30 09:23:50 - Email support at h***@m**********.org
// 2025-07-30 09:25:33 - Registered phone: ***.***.****
// 2025-07-30 09:27:05 - User j***@m***.co.uk accessed secure area
// 2025-07-30 09:29:00 - Card declined: **** **** **** 9424
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func redactNum(num string) string {
	switch {
	case strings.HasPrefix(num, "+91"):
		return "+91**********"
	case strings.HasPrefix(num, "("):
		return "(***) ***-****"
	case strings.Contains(num, "."):
		return "***.***.****"
	default:
		return ""
	}
}

func redactMail(email string) string {
	strSplit := strings.Split(email, "@")
	domain := strSplit[1]
	domainSplit := strings.Split(domain, ".")
	nonredacted := "."
	for i := 1; i < len(domainSplit)-1; i++ {
		nonredacted += domainSplit[i] + "."
	}
	nonredacted += domainSplit[len(domainSplit)-1]
	return "******@*****" + nonredacted
}

func redactcc(cc string) string {
	res := ""
	for i := 0; i < len(cc); i++ {
		num := int(cc[i]) - '0'
		if num >= 0 && num <= 9 {
			res += "*"
		} else {
			res += string(cc[i])
		}
	}
	return res
}

func getSensitiveData(line string) string {
	// mailPattern := regexp.MustCompile(`[\w\?\%\#\_\.]+@[\w\?\%\#\_\.]+.[a-zA-Z]{2,3}[\.a-zA-Z]{0,4}`)
	mailPattern := regexp.MustCompile(`[\w\?\%\#\_\.]+@[\w\?\%\#\_]+.[a-zA-Z]{2,3}[.a-z]{0,3}`)
	ccPattern := regexp.MustCompile(`[0-9]{4}(-| )[0-9]{4}(-| )[0-9]{4}(-| )([0-9]{4})`)
	numPattern := regexp.MustCompile(`(\+91\-[0-9]{10}|\([0-9]{3}\) [0-9]{3}\-[0-9]{4}|[0-9]{3}\.[0-9]{3}\.[0-9]{4})`)
	var mail, cc, num string
	mail = mailPattern.FindString(line)
	cc = ccPattern.FindString(line)
	num = numPattern.FindString(line)
	if mail != "" {
		// fmt.Println(mail, redactMail(mail))
		return strings.Replace(line, mail, redactMail(mail), 1)
	}
	if cc != "" {
		// fmt.Println(cc, redactcc(cc))
		return strings.Replace(line, cc, redactcc(cc), 1)
	}
	if num != "" {
		// fmt.Println(num, redactNum(num))
		return strings.Replace(line, num, redactcc(num), 1)
	}
	return ""
}

func main() {
	cwd, _ := os.Getwd()
	file, _ := os.Open(cwd + "/input.txt")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		res := getSensitiveData(line)
		fmt.Println(line, res)
		// fmt.Println(mail, cc, num)
	}
}
