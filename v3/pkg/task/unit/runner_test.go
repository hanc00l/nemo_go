package unit

import "testing"

func TestGetAllUnit(t *testing.T) {
	entName := ""
	cookie := ""
	var entryList []UnitEntry
	entryList = append(entryList, UnitEntry{
		CompanyName: entName,
		Depth:       Root,
	})
	err := GetAllUnit(&entryList, Root, 1, false, true, 100, cookie)
	if err != nil {
		t.Log(err)
		return
	}
	for _, entry := range entryList {
		// CompanyName 与 Company.EntName 一致表示查询正确：
		if entry.Company.EntName == entry.CompanyName {
			t.Log(entry.Depth, entry.ParentName, entry.CompanyName, entry.Company.HighlightNameType)
			for _, branch := range entry.BranchList {
				t.Log("\t", branch.BrName)
			}
			for _, invest := range entry.InvestList {
				t.Log("\t", invest.EntJgName, invest.FundedRatio)
			}
		}

	}
}
