package portscan

import (
	"testing"
)

func TestTXPortMap_ParseTxtContentResult(t *testing.T) {
	data := `192.168.120.1:53
192.168.120.1:443
192.168.120.1:80     http | [200] [Success]
192.168.120.1:8080   http | [200] [Success]
192.168.120.17:3306  mysql | 5.7.32
192.168.120.17:53
192.168.120.17:1080
192.168.120.47:10000
192.168.120.59:139   netbios-ssn
192.168.120.59:902   vmware-auth | 220 VMware Authentication Daemon Version 1.10: SSL Required, ServerDaemonProtocol:SOAP, MKSDisplayProtocol:VNC , , NFCSSL supported/t
192.168.120.59:135   msrcp
192.168.120.59:49152
192.168.120.59:49156
192.168.120.59:49154
192.168.120.59:49157
192.168.120.59:912   vmware-auth | 220 VMware Authentication Daemon Version 1.0, ServerDaemonProtocol:SOAP, MKSDisplayProtocol:VNC , ,
192.168.120.59:49153
192.168.120.59:445   microsoft-ds
192.168.120.59:49155
192.168.120.64:139   netbios-ssn
192.168.120.64:902   vmware-auth | 220 VMware Authentication Daemon Version 1.10: SSL Required, ServerDaemonProtocol:SOAP, MKSDisplayProtocol:VNC , , NFCSSL supported/t
192.168.120.64:49152
192.168.120.64:49156
192.168.120.64:49154
192.168.120.64:49157
192.168.120.64:912   vmware-auth | 220 VMware Authentication Daemon Version 1.0, ServerDaemonProtocol:SOAP, MKSDisplayProtocol:VNC , ,
192.168.120.64:135   msrcp
192.168.120.64:443   ssl/tls | [403]
192.168.120.64:49155
192.168.120.64:49153
192.168.120.64:445   microsoft-ds
192.168.120.68:62078
192.168.120.73:3306  mysql | 5.7.33
192.168.120.73:666   http | [200] [Apache/2.4.38 (Debian)|PHP/7.4.16] [码小六]
192.168.120.76:139   netbios-ssn
192.168.120.76:902   vmware-auth | 220 VMware Authentication Daemon Version 1.10: SSL Required, ServerDaemonProtocol:SOAP, MKSDisplayProtocol:VNC , , NFCSSL supported/t
192.168.120.76:1044
192.168.120.76:1026
192.168.120.76:1028
192.168.120.76:1027
192.168.120.76:1025
192.168.120.80:62078
192.168.120.116:5357 http | [503] [Microsoft-HTTPAPI/2.0] [Service Unavailable]
192.168.120.215:902  vmware-auth | 220 VMware Authentication Daemon Version 1.10: SSL Required, ServerDaemonProtocol:SOAP, MKSDisplayProtocol:VNC , , NFCSSL supported/t
192.168.120.215:5357 http | [503] [Microsoft-HTTPAPI/2.0] [Service Unavailable]
192.168.120.215:443  ssl/tls | [403]
192.168.120.215:912  vmware-auth | 220 VMware Authentication Daemon Version 1.0, ServerDaemonProtocol:SOAP, MKSDisplayProtocol:VNC , ,
192.168.120.215:7000
192.168.120.215:10001
192.168.120.215:843`
	tx := NewTXPortMap(Config{})
	tx.ParseTxtContentResult([]byte(data))
	for k, ip := range tx.Result.IPResult {
		t.Log(k)
		for kk, port := range ip.Ports {
			t.Log(kk, port.Status)
			t.Log(port.PortAttrs)
		}
	}
}
