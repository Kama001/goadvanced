// Scenario: You are monitoring application logs,
// and you need to quickly identify and extract details of specific errors.

// Problem: Write a script that reads a log file line by line.
// For any line containing an error, extract the timestamp,
// the error level (e.g., ERROR, CRITICAL, FATAL), and the specific error message.

// 2025-07-26 10:35:12,123 ERROR [main] com.example.MyService - Failed to connect to database: Connection refused
// 2025-07-26 10:36:01,456 INFO  [http-nio-8080] com.example.AnotherService - Request processed successfully
// 2025-07-26 10:37:05,789 WARN  [scheduler] com.example.TaskWorker - Task execution delayed
// 2025-07-26 10:38:15,000 CRITICAL [background] com.example.AuthService - Authentication failed for user 'johndoe': Invalid credentials

package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

func main() {
	input := `2025-07-26 10:35:12,123 ERROR [main] com.example.MyService - Failed to connect to database: Connection refused
2025-07-26 10:36:01,456 INFO  [http-nio-8080] com.example.AnotherService - Request processed successfully
2025-07-26 10:37:05,789 WARN  [scheduler] com.example.TaskWorker - Task execution delayed
2025-07-26 10:38:15,000 CRITICAL [background] com.example.AuthService - Authentication failed for user 'johndoe': Invalid credentials`
	scanner := bufio.NewScanner(strings.NewReader(input))
	ptm := regexp.MustCompile(`(\d{4}-\d{2}-\d{1,2} \d{2}:\d{2}:\d{2},\d{1,3}) (ERROR|FATAL|CRITICAL)(.*-)(.*)`)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.Contains(line, "ERROR") || strings.Contains(line, "CRITICAL") || strings.Contains(line, "FATAL") {
			matches := ptm.FindStringSubmatch(line)
			// for i := 0; i < len(matches); i++ {
			// 	fmt.Println(matches[i])
			// }
			fmt.Println("timestamp: ", matches[1], "level: ", matches[2], "message: ", matches[4])
		}
	}
}
