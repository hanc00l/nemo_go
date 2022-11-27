package notify

import "testing"

func TestNotifyMessage(t *testing.T) {
	message := "portscan->runtime:25s,runtask:0/0/9  \nresult->ip:10,port:20,domain:15,screenshot:25,vulnerability:2"
	Send(message)
}
