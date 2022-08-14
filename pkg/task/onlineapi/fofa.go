package onlineapi

// forked from https://github.com/fofapro/fofa-go

// Copyright (c) 2016 baimaohui

// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.

// Package fofa implements some fofa-api utility functions.

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Fofa a fofa client can be used to make queries
type Fofa struct {
	email []byte
	key   []byte
	*http.Client

	//Config 配置参数：查询的目标、关联的组织
	Config OnlineAPIConfig
	//Result 查询结果
	Result []onlineSearchResult
	//DomainResult 整理后的域名结果
	DomainResult domainscan.Result
	//IpResult 整理后的IP结果
	IpResult portscan.Result
}

// DomainResult represents a record of the query results
// contain domain host  ip  port title country city
type result struct {
	Domain  string `json:"domain,omitempty"`
	Host    string `json:"host,omitempty"`
	IP      string `json:"ip,omitempty"`
	Port    string `json:"port,omitempty"`
	Title   string `json:"title,omitempty"`
	Country string `json:"country,omitempty"`
	City    string `json:"city,omitempty"`
}

// User struct for fofa user
type User struct {
	Email  string `json:"email,omitempty"`
	Fcoin  int    `json:"fcoin,omitempty"`
	Vip    bool   `json:"bool,omitempty"`
	Avatar string `json:"avatar,omitempty"`
	Err    string `json:"errmsg,omitempty"`
}

// Results fofa result set
type Results []result

//

const (
	//defaultAPIUrl = "https://fofa.so/api/v1/search/all?"
	// fofa api changed:
	defaultAPIUrl = "https://fofa.info/api/v1/search/all?"
)

var (
	errFofaReplyWrongFormat = errors.New("Fofa Reply With Wrong Format")
	errFofaReplyNoData      = errors.New("No Data In Fofa Reply")
)

// NewFofaClient create a fofa client
func NewFofaClient(email, key []byte) *Fofa {

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Fofa{
		email: email,
		key:   key,
		Client: &http.Client{
			Transport: transCfg, // disable tls verify
		},
	}
}

// Get overwrite http.Get
func (ff *Fofa) Get(u string) ([]byte, error) {

	body, err := ff.Client.Get(u)
	if err != nil {
		return nil, err
	}
	defer body.Body.Close()
	content, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// QueryAsJSON make a fofa query and return json data as result
// echo 'domain="nosec.org"' | base64 - | xargs -I{}
// curl "https://fofa.info/api/v1/search/all?email=${FOFA_EMAIL}&key=${FOFA_KEY}&qbase64={}"
// host title ip domain port country city
func (ff *Fofa) QueryAsJSON(page uint, args ...[]byte) ([]byte, error) {
	var (
		query  = []byte(nil)
		fields = []byte("domain,host,ip,port,title,country,city")
		q      = []byte(nil)
	)
	switch {
	case len(args) == 1 || (len(args) == 2 && args[1] == nil):
		query = args[0]
	case len(args) == 2:
		query = args[0]
		fields = args[1]
	}

	q = []byte(base64.StdEncoding.EncodeToString(query))
	q = bytes.Join([][]byte{[]byte(defaultAPIUrl),
		[]byte("email="), ff.email,
		[]byte("&key="), ff.key,
		[]byte("&qbase64="), q,
		[]byte("&fields="), fields,
		[]byte("&page="), []byte(strconv.Itoa(int(page))),
	}, []byte(""))
	//fmt.Printf("%s\n", q)
	content, err := ff.Get(string(q))
	if err != nil {
		return nil, err
	}
	errmsg, err := jsonparser.GetString(content, "errmsg")
	if err == nil {
		err = errors.New(errmsg)
	} else {
		err = nil
	}
	return content, err
}

// QueryAsArray make a fofa query and
// return array data as result
// echo 'domain="nosec.org"' | base64 - | xargs -I{}
// curl "https://fofa.info/api/v1/search/all?email=${FOFA_EMAIL}&key=${FOFA_KEY}&qbase64={}"
func (ff *Fofa) QueryAsArray(page uint, args ...[]byte) (result Results, err error) {

	var content []byte

	content, err = ff.QueryAsJSON(page, args...)
	if err != nil {
		return nil, err
	}

	errmsg, err := jsonparser.GetString(content, "errmsg")
	// err equals to nil on error
	if err == nil {
		return nil, errors.New(errmsg)
	}
	err = json.Unmarshal(content, &result)

	return
}

// UserInfo get user information
func (ff *Fofa) UserInfo() (user *User, err error) {
	user = new(User)
	//queryStr := strings.Join([]string{"https://fofa.so/api/v1/info/my?email=", string(ff.email), "&key=", string(ff.key)}, "")
	queryStr := strings.Join([]string{"https://fofa.info/api/v1/info/my?email=", string(ff.email), "&key=", string(ff.key)}, "")

	content, err := ff.Get(queryStr)

	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(content, user); err != nil {
		return nil, err
	}

	if len(user.Err) != 0 {
		return nil, errors.New(user.Err)
	}

	return user, nil
}

func (u *User) String() string {
	data, err := json.Marshal(u)
	if err != nil {
		log.Fatalf("json marshal failed. err: %s\n", err)
		return ""
	}
	return string(data)
}

func (r *result) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		log.Fatalf("json marshal failed. err: %s\n", err)
		return ""
	}
	return string(data)
}

func (r *Results) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		log.Fatalf("json marshal failed. err: %s\n", err)
		return ""
	}
	return string(data)
}

// NewFofa 创建Fofa对象
func NewFofa(config OnlineAPIConfig) *Fofa {
	return &Fofa{Config: config}
}

// Do 执行fofa
func (ff *Fofa) Do() {
	if conf.GlobalWorkerConfig().API.Fofa.Key == "" || conf.GlobalWorkerConfig().API.Fofa.Name == "" {
		logging.RuntimeLog.Error("no fofa api key,exit fofa search")
		return
	}
	for _, line := range strings.Split(ff.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		ff.RunFofa(domain)
	}
}

// RunFofa 调用fofa搜索
func (ff *Fofa) RunFofa(domain string) {
	email := conf.GlobalWorkerConfig().API.Fofa.Name
	key := conf.GlobalWorkerConfig().API.Fofa.Key

	clt := NewFofaClient([]byte(email), []byte(key))
	if clt == nil {
		logging.RuntimeLog.Error("create fofa clien")
		return
	}
	// QueryAsJSON
	var query string
	fields := "domain,host,ip,port,title,country,city,server,banner"
	if utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
		query = fmt.Sprintf("ip=\"%s\"", domain)
	} else {
		// cert.subject相比更精准，但信息量更少；cert="xxx.com"干扰太多，暂时不用（没想法好优的方案）
		// query = fmt.Sprintf("domain=\"%s\" || host=\"%s\" || cert=\"%s\"", domain, domain, domain)
		query = fmt.Sprintf("domain=\"%s\" || cert.subject=\"%s\"", domain, domain)
	}
	// 查询第1页，并获取总共记录数量
	pageResult, sizeTotal := ff.retriedFofaSearch(clt, 1, query, fields)
	//fmt.Println(sizeTotal)
	ff.Result = append(ff.Result, pageResult...)
	// 计算需要查询的页数
	pageTotalNum := sizeTotal / pageSize
	if sizeTotal%pageSize > 0 {
		pageTotalNum++
	}
	for i := 2; i <= pageTotalNum; i++ {
		pageResult, _ = ff.retriedFofaSearch(clt, i, query, fields)
		ff.Result = append(ff.Result, pageResult...)
	}
	// 解析结果
	ff.parseResult()
}

func (ff *Fofa) retriedFofaSearch(clt *Fofa, page int, query string, fields string) (pageResult []onlineSearchResult, sizeTotal int) {
	RETRIED := 3
	for j := 0; j < RETRIED; j++ {
		ret, err := clt.QueryAsJSON(uint(page), []byte(query), []byte(fields))
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			continue
		}
		pageResult, sizeTotal = ff.parseFofaSearchResult(ret)
		if len(pageResult) > 0 {
			break
		}
	}
	return
}
