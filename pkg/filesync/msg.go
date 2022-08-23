package filesync

import "os"

type Message struct {
	MgAuthKey  string
	MgType     string //
	MgByte     []byte //
	MgString   string //
	MgStrings  []string
	MgFileMode os.FileMode
	Del        bool // whether should the not exist files in src be deleted.
	Overwrite  bool // whether the conflicted files be
}

const (
	MsgSync     = "SYNC"
	MsgMd5List  = "MD5LIST"
	MsgTran     = "TRAN"
	MsgTranData = "TRAN-DATA"
	MsgEnd      = "END"
	MsgError    = "ERROR"
)
