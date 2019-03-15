/*
Package cleanmark implements a two APIs for escaping markdown elements in raw text input.

The first API is the Normal Text API, and it revolves around the type Cleaner. It is used to convert normal text into properly formatted ubqt markdown.

The second API is the HTML API, and it revolves around the type HTMLCleaner. It is used to convert HTML into properly formatted ubqt markdown.

Normal Text API

Cleaner wraps a WriteCloser, and wraps common functions to write properly formatted markdown to the Writer - these are the *Escaped variant functions. Cleaner also provides the generic methods "Write" and "WriteString", to write unmodified text to the underlying WriteCloser.

HTML API

HTMLCleaner wraps a WriteCloser, and is used to parse html streams into their respective ubqt markdown interpretations. HTMLCleaner also provides the generic methods "Write" and "WriteString", to write unmodified text to the underlying WriteCloser.

Markdown

Ubqt-flavored markdown is described in more detail in the official document https://ubqt-systems.github.io/markdown.html 

Common markdown elements are generally easier to insert by hand,but several helper types are provided for more complex elements: color, url, and image; which are described in greater detail in greater detail below.

*/

package cleanmark



