package unit

import (
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
)

const Root = 0

type UnitEntry struct {
	CompanyName string
	ParentName  string
	Depth       int

	Company    Company
	BranchList []CompanyBranch
	InvestList []InvestCompany
}

func GetAllUnit(entryList *[]UnitEntry, depth, maxDepth int, isBranch, isInvest bool, investRatio float64, cookie string) (err error) {
	for i := range *entryList {
		e := &(*entryList)[i]
		if e.Depth == depth {
			// 获取公司信息、分支机构信息、投资机构信息
			company, branchList, investList, errUnit := GetOneUnit(e.CompanyName, isBranch, isInvest, investRatio, cookie)
			if errUnit != nil {
				return errUnit
			}
			if company != nil {
				e.Company = *company
				e.BranchList = branchList
				e.InvestList = investList
				//　下一层机构加到递归队列
				if e.Depth+1 < maxDepth {
					for _, branch := range e.BranchList {
						*entryList = append(*entryList, UnitEntry{
							CompanyName: branch.BrName,
							ParentName:  e.CompanyName,
							Depth:       e.Depth + 1,
						})
					}
					for _, invest := range e.InvestList {
						*entryList = append(*entryList, UnitEntry{
							CompanyName: invest.EntJgName,
							ParentName:  e.CompanyName,
							Depth:       e.Depth + 1,
						})
					}
				}
			}
		}
	}
	// 递归获取下一层机构
	if depth+1 < maxDepth {
		err = GetAllUnit(entryList, depth+1, maxDepth, isBranch, isInvest, investRatio, cookie)
	}

	return err
}

func GetOneUnit(companyName string, isBranch, isInvest bool, investRatio float64, cookie string) (company *Company, branchList []CompanyBranch, investList []InvestCompany, err error) {
	// 获取公司信息
	var c Company
	c, err = SearchCompany(companyName, cookie)
	if err != nil {
		logging.RuntimeLog.Errorf("获取公司信息失败：%v", err)
		return
	}
	company = &c
	// 获取webId
	var webId string
	webId, err = GetCompanyWebID(companyName, cookie)
	if err != nil {
		logging.RuntimeLog.Errorf("获取公司信息的webid失败：%v", err)
		return
	}
	if webId == "" {
		logging.RuntimeLog.Errorf("获取公司信息的webid为空")
		return
	}
	// 获取分支机构信息
	if isBranch {
		branchList, err = SearchCompanyBranch(webId, cookie)
		if err != nil {
			logging.RuntimeLog.Errorf("获取分支机构信息失败：%v", err)
		}
	}
	// 获取投资机构信息
	if isInvest {
		investList, err = SearchInvestCompanies(webId, cookie, investRatio)
		if err != nil {
			logging.RuntimeLog.Errorf("获取投资机构信息失败：%v", err)
		}
	}
	return
}
