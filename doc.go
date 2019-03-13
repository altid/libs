/*
Package cleanmark implements a two APIs for escaping markdown elements in raw text input.

The first API is the Normal Text API, and it revolves around the type Cleaner. It is used to convert normal text into properly formatted ubqt markdown.

The second API is the HTML API, and it revolves around the type HTMLCleaner. It is used to convert HTML into properly formatted ubqt markdown.

Normal Text API

Cleaner wraps a WriteCloser, and wraps common functions to write properly formatted markdown to the Writer - these are the *Escaped variant functions. Cleaner also provides the generic methods "Write" and "WriteString", to write unmodified text to the underlying WriteCloser.

HTML API

HTMLCleaner wraps a WriteCloser, and is used to parse html streams into their respective ubqt markdown interpretations. HTMLCleaner also provides the generic methods "Write" and "WriteString", to write unmodified text to the underlying WriteCloser.

Markdown

Ubqt markdown is a variant of github's markdown, with an additional element used to define colors: 

	[text](%colorcode)

For more information about github's markdown, visit https://github.github.com/gfm

*/
// A valid colorcode is anything from the following list:
/*

	red
	orange
	yellow
	green
	blue
	white
	black
	grey

*/
// 6-digit hex code (no alpha chan support currently)
/*
	Example:
	#123456
	#228822

*/
// 3-digit hex code
/*
	Example:
	#123
	#282
*/
package cleanmark



