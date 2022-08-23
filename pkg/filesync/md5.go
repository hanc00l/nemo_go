package filesync

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func Md5OfAFile(f string) (string, error) {
	fi, fiErr := os.Open(f)
	if fiErr != nil {
		// lg.Println(fiErr)
		return "", fiErr
	}
	defer fi.Close()
	r := bufio.NewReader(fi)
	h := md5.New()
	var s string
	var e error
	for {
		s, e = r.ReadString('\n')
		// lg.Println(e)
		io.WriteString(h, s)
		if e != nil {
			if e == io.EOF {
				break
			} else {
				return "", e
			}
		}
	}
	s = fmt.Sprintf("%x", h.Sum(nil))
	return s, nil
}
