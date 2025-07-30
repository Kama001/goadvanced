// Scenario: You need to automate the update of configuration files
// across multiple servers. These files often contain key-value pairs
// or structured data.

// Problem: Write a script that can find and replace a specific
// value associated with a key in a configuration file,
// without altering comments or other lines.

// Config File Example (config.properties):
// # Database settings
// db.host=localhost
// db.port=5432
// db.user=admin
// # Application settings
// app.version=1.0.0
// app.debug_mode=true

// challenge replace db.Port=5432 to db.Port=6432

package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

func main() {
	input := `# Database settings
db.host=localhost
db.port=5432
db.user=admin
# Application settings
app.version=1.0.0
app.debug_mode=true`
	res := ""
	ptm := regexp.MustCompile(`(db.port)(=)(\d{4})`)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		match := ptm.FindStringSubmatch(line)
		if len(match) > 0 {
			res += (match[1] + match[2] + "6432" + "\n")
		} else {
			res += (line + "\n")
		}
	}
	fmt.Println(res)
}
