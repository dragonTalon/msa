package model

type CmdResult struct {
	Code  int
	Msg   string
	Type  string
	Error error
	Data  interface{}
}
