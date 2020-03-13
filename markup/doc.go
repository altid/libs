/*
Package markup implements an API for escaping markdown elements in raw text input.

Cleaner wraps a WriteCloser, and wraps common functions to write properly formatted markdown to the Writer - these are the *Escaped variant functions. Cleaner also provides the generic methods "Write" and "WriteString", to write unmodified text to the underlying WriteCloser.

Altid-flavored markdown is described in more detail in the official document https://altid.github.io/markdown.html

Common markdown elements are generally easier to insert by hand,but several helper types are provided for more complex elements: color, url, and image; which are described in greater detail in greater detail below.

*/

package markup
