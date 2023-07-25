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
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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

// Domain represents a record of the query results
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
	// fofa api changed:
	defaultAPIUrl = "https://fofa.info/api/v1/search/all?"
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
func (f *Fofa) Get(u string) ([]byte, error) {
	body, err := f.Client.Get(u)
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
func (f *Fofa) QueryAsJSON(page uint, args ...[]byte) ([]byte, error) {
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
		[]byte("email="), f.email,
		[]byte("&key="), f.key,
		[]byte("&qbase64="), q,
		[]byte("&fields="), fields,
		[]byte("&size="), []byte(strconv.Itoa(pageSize)),
		[]byte("&page="), []byte(strconv.Itoa(int(page))),
	}, []byte(""))
	//fmt.Printf("%s\n", q)
	content, err := f.Get(string(q))
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
func (f *Fofa) QueryAsArray(page uint, args ...[]byte) (result Results, err error) {

	var content []byte

	content, err = f.QueryAsJSON(page, args...)
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
func (f *Fofa) UserInfo() (user *User, err error) {
	user = new(User)
	//queryStr := strings.Join([]string{"https://fofa.so/api/v1/info/my?email=", string(f.email), "&key=", string(f.key)}, "")
	queryStr := strings.Join([]string{"https://fofa.info/api/v1/info/my?email=", string(f.email), "&key=", string(f.key)}, "")

	content, err := f.Get(queryStr)

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
func (f *Fofa) Do() {
	if conf.GlobalWorkerConfig().API.Fofa.Key == "" || conf.GlobalWorkerConfig().API.Fofa.Name == "" {
		logging.RuntimeLog.Error("no fofa api key,exit fofa search")
		return
	}
	blackDomain := custom.NewBlackDomain()
	blackIP := custom.NewBlackIP()
	for _, line := range strings.Split(f.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if utils.CheckIPV4(domain) && blackIP.CheckBlack(domain) {
			continue
		}
		if utils.CheckDomain(domain) && blackDomain.CheckBlack(domain) {
			continue
		}
		f.RunFofa(domain)
	}
}

// RunFofa 调用fofa搜索
func (f *Fofa) RunFofa(domain string) {
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
	if f.Config.SearchByKeyWord {
		query = f.Config.Target
	} else {
		if utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
			query = fmt.Sprintf("ip=\"%s\"", domain)
		} else {
			// cert.subject相比更精准，但信息量更少；cert="xxx.com"干扰太多，暂时不用（没想法好优的方案）
			// 在域名前加.减少模糊匹配带来的部份干扰
			domainCert := domain
			if strings.HasPrefix(domain, ".") == false {
				domainCert = "." + domain
			}
			query = fmt.Sprintf("domain=\"%s\" || cert=\"%s\" || cert.subject=\"%s\"", domain, domainCert, domainCert)
		}
	}
	if f.Config.IsIgnoreOutofChina {
		query = fmt.Sprintf("(%s) && country=\"CN\" && region!=\"HK\" && region!=\"TW\"  && region!=\"MO\"", query)
	}
	// 查询第1页，并获取总共记录数量
	pageResult, sizeTotal := f.retriedFofaSearch(clt, 1, query, fields)
	if f.Config.SearchLimitCount > 0 && sizeTotal > f.Config.SearchLimitCount {
		sizeTotal = f.Config.SearchLimitCount
	}
	//fmt.Println(sizeTotal)
	f.Result = append(f.Result, pageResult...)
	// 计算需要查询的页数
	pageTotalNum := sizeTotal / pageSize
	if sizeTotal%pageSize > 0 {
		pageTotalNum++
	}
	for i := 2; i <= pageTotalNum; i++ {
		pageResult, _ = f.retriedFofaSearch(clt, i, query, fields)
		f.Result = append(f.Result, pageResult...)
	}
	// 解析结果
	f.parseResult()
}

func (f *Fofa) retriedFofaSearch(clt *Fofa, page int, query string, fields string) (pageResult []onlineSearchResult, sizeTotal int) {
	RETRIED := 3
	for j := 0; j < RETRIED; j++ {
		ret, err := clt.QueryAsJSON(uint(page), []byte(query), []byte(fields))
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			if strings.Contains(err.Error(), "F点余额不足") || strings.Contains(err.Error(), "请按顺序进行翻页查询") {
				break
			}
			if strings.Contains(err.Error(), "请求速度过快") {
				time.Sleep(2 * time.Second)
			}
			continue
		}
		//fmt.Println(string(ret))
		pageResult, sizeTotal = f.parseFofaSearchResult(ret)
		if len(pageResult) > 0 {
			break
		}
	}
	return
}

// ParseCSVContentResult 解析导出的CSV文本结果
func (f *Fofa) ParseCSVContentResult(content []byte) {
	s := custom.NewService()
	if f.IpResult.IPResult == nil {
		f.IpResult.IPResult = make(map[string]*portscan.IPResult)
	}
	if f.DomainResult.DomainResult == nil {
		f.DomainResult.DomainResult = make(map[string]*domainscan.DomainResult)
	}
	r := csv.NewReader(bytes.NewReader(content))
	for index := 0; ; index++ {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		//忽略第一行的标题行
		if err != nil || index == 0 {
			continue
		}
		domain := utils.HostStrip(strings.TrimSpace(row[0]))
		ip := strings.TrimSpace(row[2])
		port, portErr := strconv.Atoi(row[3])
		title := strings.TrimSpace(row[4])
		service := strings.TrimSpace(row[6])
		//域名属性：
		if len(domain) > 0 && utils.CheckIPV4(domain) == false {
			if f.DomainResult.HasDomain(domain) == false {
				f.DomainResult.SetDomain(domain)
			}
			if len(ip) > 0 {
				f.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "fofa",
					Tag:     "A",
					Content: ip,
				})
			}
			if len(title) > 0 {
				f.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "fofa",
					Tag:     "title",
					Content: title,
				})
			}
		}
		//IP属性（由于不是主动扫描，忽略导入StatusCode）
		if len(ip) == 0 || utils.CheckIPV4(ip) == false || portErr != nil {
			continue
		}
		if f.IpResult.HasIP(ip) == false {
			f.IpResult.SetIP(ip)
		}
		if f.IpResult.HasPort(ip, port) == false {
			f.IpResult.SetPort(ip, port)
		}
		if len(title) > 0 {
			f.IpResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "fofa",
				Tag:     "title",
				Content: title,
			})
		}
		if len(service) <= 0 || service == "unknown" {
			service = s.FindService(port, "")
		}
		if len(service) > 0 {
			f.IpResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "fofa",
				Tag:     "service",
				Content: service,
			})
		}
	}
}
