package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	maxRetries      = 3
	retryDelay      = 1 * time.Second
	requestTimeout  = 30 * time.Second
	defaultPageSize = 100
	userAgent       = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.71 Safari/537.36"
)

// Company represents a single company entry in the response
type Company struct {
	EntName           string `json:"entName"`
	ENTSTATUS         string `json:"ENTSTATUS"`
	Entid             string `json:"entid"`
	DataType          string `json:"dataType"`
	ENTNAME           string `json:"ENTNAME"`
	EntStatus         string `json:"entStatus"`
	HighlightNameType string `json:"highlightNameType"`
}

// CompanyBranch represents the apiData structure
type CompanyBranch struct {
	BrName         string `json:"brName"`
	BrPrincipal    string `json:"brPrincipal"`
	EsDate         string `json:"esDate"`
	EntStatus      string `json:"ent_status"`
	BrnCreditCode  string `json:"brnCreditCode"`
	BrnRegOrg      string `json:"brnRegOrg"`
	BrRegNo        string `json:"brRegNo"`
	Entid          string `json:"entid"`
	PersonId       string `json:"personId"`
	PersonEntCount int    `json:"personEntCount"`
}

// InvestCompany represents the company information in the apiData array
type InvestCompany struct {
	EntJgName   string `json:"entJgName"`
	EntStatus   string `json:"entStatus"`
	EntType     string `json:"entType"`
	EsDate      string `json:"esDate"`
	FundedRatio string `json:"fundedRatio"`
	Name        string `json:"name"`
	RegCap      string `json:"regCap"`
	RegCapCur   string `json:"regCapCur"`
	RegNo       string `json:"regNo"`
	RegOrg      string `json:"regOrg"`
	RegOrgCode  string `json:"regOrgCode"`
	RevDate     string `json:"revDate"`
	SubConAm    string `json:"subConAm"`
	Entid       string `json:"entid"`
	Personid    string `json:"personid"`
}

// CompanyAPIResponse represents the structure of the API response
type CompanyAPIResponse struct {
	Code    int       `json:"code"`
	Msg     string    `json:"msg"`
	Data    []Company `json:"data"`
	Success bool      `json:"success"`
}

// CompanyBranchAPIResponse represents the overall response structure
type CompanyBranchAPIResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Data    struct {
		APIData []CompanyBranch `json:"apiData"`
	} `json:"data"`
}

// InvestCompanyApiResponse represents the overall API response structure
type InvestCompanyApiResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Data    struct {
		APIData []InvestCompany `json:"apiData"`
	} `json:"data"`
}

// SearchCompany searches for companies matching the key
func SearchCompany(companyName, cookie string) (company Company, error error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// URL encode the companyName
	encodedKey := url.QueryEscape(companyName)

	// Create the request URL
	apiUrl := fmt.Sprintf("https://www.riskbird.com/riskbird-api/searchHint?queryType=3&key=%s", encodedKey)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return company, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("app-device", "WEB")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("https://www.riskbird.com/search/company?keyword=%s&timestamp=%d", encodedKey, time.Now().UnixNano()/int64(time.Millisecond)))
	req.Header.Set("Cookie", cookie)

	// Configure HTTP client with timeout and retry logic
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var resp *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return company, ctx.Err()
			}
		}

		resp, err = client.Do(req)
		if err == nil {
			break
		}

		if attempt == maxRetries-1 {
			return company, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
		}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return company, fmt.Errorf("failed to read response body: %w", err)
	}
	// Parse JSON response
	var apiResponse CompanyAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return company, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	// Filter entries where entName exactly matches the companyName
	for _, entry := range apiResponse.Data {
		if entry.EntName == companyName {
			return entry, nil
		}
	}

	return company, fmt.Errorf("failed to find entry for companyName: %s", companyName)
}

func GetCompanyWebID(entName, cookie string) (string, error) {
	// 编码公司名称
	encodedEntName := url.QueryEscape(entName)
	apiUrl := fmt.Sprintf("https://www.riskbird.com/ent/%s.html", encodedEntName)

	// 准备请求
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Cookie", cookie)

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: requestTimeout,
	}

	var resp *http.Response
	var lastErr error
	// 重试逻辑
	for i := 0; i < maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
	}

	if resp == nil {
		return "", fmt.Errorf("请求失败，最后错误: %v", lastErr)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %v", err)
	}
	//查询次数已达到上限
	if strings.Contains(string(body), "查询次数已达到上限") {
		return "", fmt.Errorf("您已达到查询次数上限，请稍后再试")
	}
	// 使用更灵活的正则表达式提取WEB开头的ID
	re := regexp.MustCompile(`WEB\d+`)
	match := re.FindString(string(body))
	if match == "" {
		return "", fmt.Errorf("未找到匹配的WEB ID")
	}

	return match, nil
}

// SearchCompanyBranch searches for companies with pagination and retry logic
func SearchCompanyBranch(orderNo, cookie string) ([]CompanyBranch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	var allResults []CompanyBranch
	page := 1

	for {
		// Make request with retry logic
		apiData, err := searchCompanyBranchPageWithRetry(ctx, orderNo, cookie, page)
		if err != nil {
			return nil, fmt.Errorf("failed to search company page %d: %w", page, err)
		}

		// Append results
		// 只保留在营公司
		for _, entry := range apiData {
			if entry.EntStatus == "在营" {
				allResults = append(allResults, entry)
			}
		}
		//allResults = append(allResults, apiData...)

		// Stop pagination if we get an empty page
		if len(apiData) == 0 {
			break
		}

		page++
	}

	return allResults, nil
}

// searchCompanyBranchPageWithRetry makes the actual HTTP request with retry logic
func searchCompanyBranchPageWithRetry(ctx context.Context, orderNo, cookie string, page int) ([]CompanyBranch, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
				// Wait before retry
			}
		}

		apiData, err := searchCompanyBranchPage(ctx, orderNo, cookie, page)
		if err == nil {
			return apiData, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("after %d retries, last error: %w", maxRetries, lastErr)
}

// searchCompanyBranchPage makes a single page request
func searchCompanyBranchPage(ctx context.Context, orderNo, cookie string, page int) ([]CompanyBranch, error) {
	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// Build URL
	apiUrl := fmt.Sprintf("https://www.riskbird.com/riskbird-api/query/company/searchCompanyPageByWeb?page=%d&size=%d&orderNo=%s&extractType=fzjg",
		page,
		defaultPageSize,
		orderNo,
	)

	req, err := http.NewRequestWithContext(reqCtx, "GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", cookie)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse CompanyBranchAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("API error: %s (code: %d)", apiResponse.Msg, apiResponse.Code)
	}

	return apiResponse.Data.APIData, nil
}

// SearchInvestCompanies retrieves all companies by paginating through results
func SearchInvestCompanies(orderNo, cookie string, rating float64) ([]InvestCompany, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	client := &http.Client{
		Timeout: requestTimeout,
	}

	var allCompanies []InvestCompany
	page := 1

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue
		}

		// Build URL with query parameters
		u, err := url.Parse("https://www.riskbird.com/riskbird-api/query/company/searchCompanyPageByWeb")
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}

		q := u.Query()
		q.Add("page", fmt.Sprintf("%d", page))
		q.Add("size", fmt.Sprintf("%d", defaultPageSize))
		q.Add("orderNo", orderNo)
		q.Add("extractType", "qydwtz")
		u.RawQuery = q.Encode()

		var lastErr error
		var resp *http.Response

		// Retry logic
		for attempt := 0; attempt < maxRetries; attempt++ {
			if attempt > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Second * time.Duration(attempt)):
					// Exponential backoff
				}
			}

			// Create request
			req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
			if err != nil {
				lastErr = fmt.Errorf("failed to create request: %w", err)
				continue
			}

			// Set headers
			req.Header.Set("User-Agent", userAgent)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Cookie", cookie)

			// Execute request
			resp, err = client.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("request failed: %w", err)
				continue
			}
			defer resp.Body.Close()

			// Read response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				lastErr = fmt.Errorf("failed to read response: %w", err)
				continue
			}

			// Parse response
			var apiResp InvestCompanyApiResponse
			if err := json.Unmarshal(body, &apiResp); err != nil {
				lastErr = fmt.Errorf("failed to parse response: %w", err)
				continue
			}

			if !apiResp.Success || apiResp.Code != 20000 {
				lastErr = fmt.Errorf("API error: %s (code: %d)", apiResp.Msg, apiResp.Code)
				continue
			}
			// 只保留在营公司
			for _, company := range apiResp.Data.APIData {
				if company.EntStatus == "在营" {
					// 检查百分比
					percent, err := ParsePercentage(company.FundedRatio)
					if err != nil {
						lastErr = fmt.Errorf("failed to parse percentage: %w", err)
						continue
					}
					if percent >= rating {
						allCompanies = append(allCompanies, company)
					}
				}
			}
			//allCompanies = append(allCompanies, apiResp.Data.APIData...)

			// Check if we should continue paginating
			if len(apiResp.Data.APIData) < defaultPageSize {
				return allCompanies, nil
			}

			// Success - break out of retry loop
			break
		}

		if lastErr != nil {
			return nil, fmt.Errorf("after %d attempts, last error: %w", maxRetries, lastErr)
		}

		page++
	}
}

func ParsePercentage(percentStr string) (float64, error) {
	numberStr := strings.TrimSuffix(percentStr, "%")
	value, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析百分比: %v", err)
	}
	return value, nil
}
