/*Package files is a layer over a normal directory to synthesize Stat responses, and allow special semantics for Read/Write requests to opened files

Adding

	import (
		"os"
		"path"

		"github.com/altid/server/files"
	)

	type NormalHandler struct{}

	func NewNormal() *NormalHandler { return &NormalHandler{} }

	func (*NormalHandler) Normal(msg *files.Message) (interface{}, error) {
		fp := path.Join(msg.Service, msg.Buffer, msg.Target)
		return os.Open(fp)
	}

	func (*NormalHandler) Stat(msg *files.Message) (os.FileInfo, error) {
		fp := path.Join(msg.Service, msg.Buffer, msg.Target)
		return os.Lstat(fp)
	}

	func main() {
		fh := files.Handle("/path/to/dir")
		fh.Add("/myfile", NewNormal())

		// [...]
	}
*/
package files
