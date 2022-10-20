package poclib

import (
	"regexp"
	"strings"
)

type RuleStruct struct {
	Opt     string
	Rules   [][]string
	SubRule []RuleStruct
}

type Content struct {
	App      string
	Port     string
	Body     string
	Header   string
	Banner   string
	Server   string
	Title    string
	Protocol string
	Product  string
	Cert     string
}

func (c Content) GetInfo(s1 string) string {
	switch s1 {
	case "app":
		return c.App
	case "port":
		return c.Port
	case "body":
		return c.Body
	case "header":
		return c.Header
	case "banner":
		return c.Banner
	case "server":
		return c.Server
	case "title":
		return c.Title
	case "protocol":
		return c.Protocol
	case "product":
		return c.Product
	case "cert":
		return c.Cert
	default:
		return ""
	}
}

func InCollections(s1 []string, s2 string) bool {
	for _, s := range s1 {
		if s == s2 {
			return true
		}
	}
	return false
}

// 获取单个规则属性
func GetOneRule(ruletxt string) []string {
	result := []string{}
	//RuleOpt := []string{"==", "=", "!="}
	//ruleOpts := strings.Join(RuleOpt, "|")
	re_rule := "(==|=|!=)+"

	reg, err := regexp.Compile(re_rule)
	if err != nil {
		return result
	}
	result1 := reg.FindStringIndex(ruletxt)
	if len(result1) == 2 {
		result = []string{ruletxt[:result1[0]], ruletxt[result1[0]:result1[1]], ruletxt[result1[1]:]}
	}
	//println(result)
	return result
}
func MatchRuleOpt(rule []string, content Content) bool {
	opts := []string{"=", "==", "!="}
	keys := []string{"app", "port", "body", "header", "banner", "server", "title", "product", "protocol", "cert"}
	if len(rule) == 3 && InCollections(keys, rule[0]) && InCollections(opts, rule[1]) {
		opt := rule[1]
		key := rule[0]
		rule_str := rule[2]
		if len(rule_str) > 2 && rule_str[0] == '"' && rule_str[len(rule_str)-1] == '"' {
			rule_str = strings.Trim(rule_str[1:len(rule_str)-1], " ")
		}
		if (opt == "==" || opt == "=") && (strings.Contains(content.GetInfo(key), rule_str) || strings.Contains(strings.ToLower(content.GetInfo(key)), strings.ToLower(rule_str))) {
			return true
		} else if (opt == "!=") && content.GetInfo(key) != "" && !strings.Contains(content.GetInfo(key), rule_str) && !strings.Contains(strings.ToLower(content.GetInfo(key)), strings.ToLower(rule_str)) {
			return true
		}
	}
	return false
}

func MatchRules(rules RuleStruct, content Content) bool {
	default_result := false
	defer func() {
		if err := recover(); err != nil {
			default_result = false
		}
	}()
	if len(rules.Rules) > 0 || len(rules.SubRule) > 0 {
		opt := "OR"
		if rules.Opt == "AND" {
			opt = "AND"
			default_result = true
		}
		for _, rule := range rules.Rules {
			if opt == "OR" {
				if MatchRuleOpt(rule, content) {
					return true
				}
			} else {
				if !MatchRuleOpt(rule, content) {
					return false
				}
			}
		}
		if len(rules.SubRule) > 0 {
			for _, subRule := range rules.SubRule {
				if opt == "OR" {
					if MatchRules(subRule, content) {
						return true
					}
				} else {
					if !MatchRules(subRule, content) {
						return false
					}
				}
			}
		}
	}
	return default_result
}

type KhIndex struct {
	L int
	R int
}

func GetKhtxt(rules_txt string) []string {
	inx_l := []int{}
	inx_r := []int{}
	inx_x := []KhIndex{}
	for i := 0; i < len(rules_txt); i++ {
		if '(' == rules_txt[i] {
			inx_l = append(inx_l, i)
			inx_x = append(inx_x, KhIndex{0, i})
		} else if ')' == rules_txt[i] {
			inx_r = append(inx_r, i)
			inx_x = append(inx_x, KhIndex{1, i})
		}
	}

	lastIndex := 0
	KhTxt := []string{}

	for indexi, nx := range inx_x {
		if nx.L == 0 && nx.R >= lastIndex {
			ix := 0
			for ii := indexi + 1; ii < len(inx_x); ii++ {
				if inx_x[ii].L == 0 {
					ix = ix + 1
				} else if inx_x[ii].L == 1 {
					if ix == 0 {
						lastIndex = inx_x[ii].R
						strxyz := rules_txt[nx.R : inx_x[ii].R+1]
						if strxyz != rules_txt {
							KhTxt = append(KhTxt, strxyz)
						}
						break
					}
					ix = ix - 1
				}
			}
		}
	}
	return KhTxt
}

func ParseRules(ruleTxt string) *RuleStruct {

	ruleTxt = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(ruleTxt, " || ", "||"), " && ", "&&"), " | ", "||")
	ruleStruct := RuleStruct{SubRule: []RuleStruct{}}

	//result := [][]string{}
	if strings.Contains(ruleTxt, "||") && !strings.Contains(ruleTxt, "&&") && (!strings.Contains(ruleTxt, "(") || !strings.Contains(ruleTxt, ")")) {
		ruleStruct.Opt = "OR"
		ortxts := strings.Split(ruleTxt, "||")
		for _, ortxt := range ortxts {
			ruleOne := GetOneRule(ortxt)
			if len(ruleOne) == 3 {
				ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
			}
		}
	} else if !strings.Contains(ruleTxt, "&&") && !strings.Contains(ruleTxt, "||") {
		ruleStruct.Opt = "AND"
		ruleOne := GetOneRule(ruleTxt)
		if len(ruleOne) == 3 {
			ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
		}
	} else if strings.Contains(ruleTxt, "&&") && !strings.Contains(ruleTxt, "||") && (!strings.Contains(ruleTxt, "(") || !strings.Contains(ruleTxt, ")")) {
		ruleStruct.Opt = "AND"
		ortxts := strings.Split(ruleTxt, "&&")
		for _, ortxt := range ortxts {
			ruleOne := GetOneRule(ortxt)
			if len(ruleOne) == 3 {
				ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
			}
		}
	} else if strings.Contains(ruleTxt, "&&") && strings.Contains(ruleTxt, "||") && (!strings.Contains(ruleTxt, "(") || !strings.Contains(ruleTxt, ")")) {
		ruleStruct.Opt = "OR"
		ortxts := strings.Split(ruleTxt, "||")
		for _, ortxt := range ortxts {
			if strings.Contains(ortxt, "&&") {
				subRule := ParseRules(ortxt)
				ruleStruct.SubRule = append(ruleStruct.SubRule, *subRule)
			} else {
				ruleOne := GetOneRule(ortxt)
				if len(ruleOne) == 3 {
					ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
				}
			}
		}
	} else if strings.Contains(ruleTxt, "(") {
		if !strings.Contains(ruleTxt, "||") && strings.Contains(ruleTxt, "&&") {
			ruleStruct.Opt = "AND"
			ortxts := strings.Split(ruleTxt, "&&")
			for _, ortxt := range ortxts {
				ruleOne := GetOneRule(ortxt)
				if len(ruleOne) == 3 {
					ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
				}
			}
		} else if !strings.Contains(ruleTxt, "||") && !strings.Contains(ruleTxt, "&&") {
			ruleStruct.Opt = "OR"
			ortxts := strings.Split(ruleTxt, "||")
			for _, ortxt := range ortxts {
				ruleOne := GetOneRule(ortxt)
				if len(ruleOne) == 3 {
					ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
				}
			}
		} else {
			khtxt := GetKhtxt(ruleTxt)
			pyx := ruleTxt
			if len(khtxt) > 0 {
				for _, khx := range khtxt {
					pyx = strings.ReplaceAll(pyx, khx, "?_________TXT__?_TXT__________?")
				}
				if strings.Contains(pyx, "||") && !strings.Contains(pyx, "&&") {
					ruleStruct.Opt = "OR"
					ortxts := strings.Split(pyx, "||")
					indexxx := 0
					for _, ortxt := range ortxts {
						if ortxt == "?_________TXT__?_TXT__________?" {
							rul4 := ParseRules(khtxt[indexxx][1 : len(khtxt[indexxx])-1])
							ruleStruct.SubRule = append(ruleStruct.SubRule, *rul4)
							indexxx++
						} else {
							ruleOne := GetOneRule(ortxt)
							if len(ruleOne) == 3 {
								ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
							}
						}

					}

				} else if !strings.Contains(pyx, "||") && strings.Contains(pyx, "&&") {
					ruleStruct.Opt = "AND"
					ortxts := strings.Split(pyx, "&&")
					indexxx := 0
					for _, ortxt := range ortxts {
						if ortxt == "?_________TXT__?_TXT__________?" {
							rul4 := ParseRules(khtxt[indexxx][1 : len(khtxt[indexxx])-1])
							ruleStruct.SubRule = append(ruleStruct.SubRule, *rul4)
							indexxx++
						} else {
							ruleOne := GetOneRule(ortxt)
							if len(ruleOne) == 3 {
								ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
							}
						}

					}
				} else if strings.Contains(pyx, "||") && strings.Contains(pyx, "&&") {
					ruleStruct.Opt = "OR"
					ortxts := strings.Split(pyx, "||")
					indexxx := 0
					for _, ortxt := range ortxts {
						if strings.Contains(ortxt, "?_________TXT__?_TXT__________?") {
							if strings.Contains(ortxt, "&&") {
								rul4 := ParseRules(strings.ReplaceAll(ortxt, "?_________TXT__?_TXT__________?", khtxt[indexxx]))
								ruleStruct.SubRule = append(ruleStruct.SubRule, *rul4)
								indexxx++
							} else {
								rul4 := ParseRules(khtxt[indexxx][1 : len(khtxt[indexxx])-1])
								ruleStruct.SubRule = append(ruleStruct.SubRule, *rul4)
								indexxx++
							}
						} else {
							ruleOne := GetOneRule(ortxt)
							if len(ruleOne) == 3 {
								ruleStruct.Rules = append(ruleStruct.Rules, ruleOne)
							}
						}

					}
				}
			} else {
				if strings.HasPrefix(ruleTxt, "(") && strings.HasSuffix(ruleTxt, ")") {
					return ParseRules(ruleTxt[1 : len(ruleTxt)-1])
				}
			}
		}
	}
	return &ruleStruct
}
