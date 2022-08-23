package filesync

import (
	"os"
)

func localOP(slinkNeedCreat map[string]string, slinkNeedChange map[string]string, needDelete []string, needCreDir []string) error {
	var err error
	for _, v := range needDelete {
		err = os.RemoveAll(v)
		if err != nil {
			DebugInfor(err)
			return err
		}
	}
	for _, v := range needCreDir {
		err = os.MkdirAll(v, 0755)
		if err != nil {
			DebugInfor(err)
			return err
		}
	}
	for k, v := range slinkNeedCreat {
		// err = os.Symlink(k, v)
		err = os.Symlink(v, k)
		if err != nil {
			DebugInfor(err)
			return err
		}
	}
	for k, v := range slinkNeedChange {
		err = os.Remove(k)
		if err != nil {
			DebugInfor(err)
			return err
		}
		// err = os.Symlink(k, v)
		err = os.Symlink(v, k)
		if err != nil {
			DebugInfor(err)
			return err
		}
	}
	return nil
}
