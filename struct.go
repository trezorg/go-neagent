package main

type neagentArgs struct {
	Database string
	File     string
	Link     string
	Config   string
	Bot      string
	Cid      string
	Timeout  int
	Verbose  bool
	Daemon   bool
	Stdout   bool
	Telegram bool
}

type strResult struct {
	error  error
	result string
}
