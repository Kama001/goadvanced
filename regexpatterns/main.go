package main

import (
	"fmt"
	"regexp"
)

func main() {
	// ptm := regexp.MustCompile("in")
	// if ptm.MatchString("this sin sub") {
	// 	fmt.Println("found")
	// }

	// ?i is for case insensitive match
	// \b will match exact word, if there are apples it wont match
	// \w matches A-Z, a-z, 0-9, _
	// \w* means after apple there can be other chars
	// but after that word ends match will stop
	// ===============================================
	// text := "Apples5 are delicious, but so are apples, aPpleS, and APPlE."
	// ptmtext := regexp.MustCompile(`\b(?i)apple\w*\b`)
	// matchesText := ptmtext.FindAllString(text, -1)
	// fmt.Println(matchesText)

	// (?m) This is especially useful when you have a multiline text
	// and want to apply the pattern to each line separately.
	// ================================================
	// text1 := "Hello World\nHello Golang\nHello Regexp"
	// ptmtext1 := regexp.MustCompile(`(?m)^Hello`)
	// matchesText1 := ptmtext1.FindAllString(text1, -1)
	// fmt.Println(matchesText1)

	// text2 := "The price is $20, the code is ABC123."
	// ptmtext2 := regexp.MustCompile(`\w+\d\w*`)
	// matchesText2 := ptmtext2.FindAllString(text2, -1)
	// fmt.Println(matchesText2)

	// ^ symbol
	// At beginning of regex -> Match the start of the string
	// Inside [] at start -> Match any character NOT in the set
	// =========================================================
	// text3 := "Hello, how are you?"
	// ptmtext3 := regexp.MustCompile(`[^aeiouAEIOU]`)
	// matchesText3 := ptmtext3.FindAllString(text3, -1)
	// fmt.Println(matchesText3)

	// ranges
	// ===================================================
	// text4 := "The quick brown fox jumps over the lazy dog."

	// // Match any lowercase letter from 'a' to 'z'
	// ptmtext4 := regexp.MustCompile("[a-z]")
	// lowercaseLetters := ptmtext4.FindAllString(text4, -1)
	// fmt.Println("Lowercase letters:", lowercaseLetters)

	// // Match any word character (alphanumeric + underscore)
	// ptmtext5 := regexp.MustCompile("[A-Za-z0-9_]")
	// wordCharacters := ptmtext5.FindAllString(text4, -1)
	// fmt.Println("Word characters:", wordCharacters)

	// // Match any character that is NOT a space
	// ptmtext6 := regexp.MustCompile("[^ ]")
	// nonSpaceCharacters := ptmtext6.FindAllString(text4, -1)
	// fmt.Println("Non-space characters:", nonSpaceCharacters)

	// Curly Braces ({min,max})
	// ================================================
	ptmtext7 := regexp.MustCompile(`go{2,4}`)
	fmt.Println(ptmtext7.FindAllString("go", -1))
	fmt.Println(ptmtext7.FindAllString("goo", -1))
	fmt.Println(ptmtext7.FindAllString("goooooooo", -1)) // op: goooo

	// check this page for more examples
	// https://www.honeybadger.io/blog/a-definitive-guide-to-regular-expressions-in-go/
}
