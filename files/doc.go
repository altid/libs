/*Package files is a layer over a normal directory to synthesize Stat responses, and allow special semantics for Read/Write requests to opened files

Adding

To add a special file, simply include it in this source directory, making sure to call the AddFileHandler with your specific handlers in the init()

	func init() {
		s := &files.Handler{
			Normal: getMyFile,
			Stat:   getMyFileStat,
		}
		files.Add("/myfile", s)
	}

	func getMyFile(msg *files.Message) (interface{}, error) {
		fp := path.Join(*inpath, msg.service, msg.buff, msg.file)
		return os.Open(fp)
	}

	func getMyFileStat (msg *files.Message) (os.FileInfo, error) {
		fp := path.Join(*inpath, msg.service, msg.buff, msg.file)
		return os.Lstat(fp)
	}
*/
package files