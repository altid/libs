package fs

type Ctrl interface {
	Open(filename string)
	Close(filename string)
	Default(filename string)
}

type Control struct {
	tabs map[string]string
	logDir string
	ctrlDir string
	mtpt string
	req chan string
	done chan struct{}
	ctrl *Ctrl
}

// NewCtrlFile returns a ready to go Control
func NewCtrlFile(ctrl *Ctrl, logDir, ctrlDir string) (*Control, error) {
	tab := make(map[string]string)
	req := make(chan string)
	done := make(chan struct{})
	control := &Control{
		tabs: tab,
		req: req,
		done: done,
		ctrlDir: ctrlDir,
		logDir: logDir,
		ctrl: ctrl
	}
}

func (c *Control) Cleanup() {
	os.RemoveAll(c.mtpt)
}

func (c *Control) Listen() {
	go dispatch(c, logdir, ctrldir)
	r, err := NewReader(path.Join(c.ctrlDir + "ctrl"))
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "quit" {
			close(c.done)
	 		break
		}
		c.req <- line
	}
}

func dispatch(c *Control, logdir, ctrldir) {
	// TODO:
	// If close is requested on a file which is currently being opened, cancel open request
	// If open is requested on file which already exists, no-op
	for {
		select {
		case line := <- c.req:
			token := strings.Fields(line)
			switch token[0] {
			case "open":
				err := createDir(logdir, ctrldir)
				if err != nil {
					return nil, err
				}
				c.Tab[token[1] = token[1]
				c.Ctrl.Open(token[1])
			case "close":
				delete(c.Tab, token[1])
				c.Ctrl.Close(token[1])
				os.RemoveAll(ctrldir)
			default:
				c.Ctrl.Default(line)
		case <- c.done:
			break
		}
}

func createDir(mtpt, logdir, ctrldir string) error {
	if _, err := os.Stat(logdir); os.IsNotExist(err) {
		fp, _ := os.Create(logfile)
		fp.Close()
	}
	if runtime.GOOS == "plan9" {
		command := exec.Command("/bin/bind", logfile, ctrldir)
		return command.Run()
	}
	return os.Symlink(logdir, ctrldir)
}
