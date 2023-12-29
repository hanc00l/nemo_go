package portscan

import (
	"testing"
)

func TestGogo_Run(t *testing.T) {
	config := Config{
		Target:        "127.0.0.1,172.16.222.1",
		ExcludeTarget: "",
		Port:          "3306,1080,4369,4572",
		Rate:          1000,
		IsPing:        false,
	}
	g := NewGogo(config)
	g.Do()
	t.Log(&g.Result)
	for ip, ipa := range g.Result.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}

var gogoResult = `{
    "config": {
        "ip": "192.168.3.1/24",
        "ips": null,
        "ports": "common",
        "json_file": "",
        "list_file": "",
        "threads": 4000,
        "mod": "default",
        "alive_spray": null,
        "port_spray": false,
        "exploit": "auto",
        "json_type": "scan",
        "version_level": 1
    },
    "data": [
        {
            "ip": "192.168.3.1",
            "port": "22",
            "frameworks": {
                "ssh": {
                    "name": "ssh",
                    "froms": {
                        "0": true
                    }
                }
            },
            "protocol": "tcp",
            "status": "tcp",
            "language": "",
            "title": "SSH-2.0-dropb",
            "midware": ""
        },
        {
            "ip": "192.168.3.1",
            "port": "8443",
            "host": "router.asus.com",
            "protocol": "tcp",
            "status": "200",
            "language": "",
            "title": "HTTP/1.0 200",
            "midware": "httpd/2.0"
        },
        {
            "ip": "192.168.3.1",
            "port": "80",
            "protocol": "http",
            "status": "200",
            "language": "",
            "title": "HTTP/1.0 200",
            "midware": "httpd/2.0"
        },
        {
            "ip": "192.168.3.15",
            "port": "8888",
            "protocol": "http",
            "status": "200",
            "language": "",
            "title": "HTTP/1.1 200",
            "midware": ""
        },
        {
            "ip": "192.168.3.107",
            "port": "1080",
            "frameworks": {
                "socks5": {
                    "name": "socks5",
                    "froms": {
                        "1": true
                    }
                }
            },
            "vulns": [
                {
                    "name": "socks5_unauthorized",
                    "severity": 3
                }
            ],
            "protocol": "tcp",
            "status": "tcp",
            "language": "",
            "title": "",
            "midware": ""
        },
        {
            "ip": "192.168.3.107",
            "port": "15672",
            "frameworks": {
                "rabbitmq-manager": {
                    "name": "rabbitmq-manager",
                    "froms": {
                        "0": true,
                        "2": true
                    },
                    "tags": [
                        "cloud"
                    ]
                }
            },
            "protocol": "http",
            "status": "200",
            "language": "",
            "title": "RabbitMQ Management",
            "midware": "Cowboy"
        },
        {
            "ip": "192.168.3.241",
            "port": "22",
            "frameworks": {
                "ssh": {
                    "name": "ssh",
                    "froms": {
                        "0": true
                    }
                }
            },
            "protocol": "tcp",
            "status": "tcp",
            "language": "",
            "title": "SSH-2.0-dropb",
            "midware": ""
        },
        {
            "ip": "192.168.3.241",
            "port": "80",
            "protocol": "http",
            "status": "200",
            "language": "",
            "title": "HTTP/1.0 200",
            "midware": "httpd/2.0"
        }
    ]
}`

func TestGogo_ParseJsonContentResult(t *testing.T) {
	i := NewImportOfflineResult("gogo")
	i.Parse([]byte(gogoResult))
	for ip, ipa := range i.IpResult.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}
