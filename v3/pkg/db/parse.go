package db

import (
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"strconv"
	"strings"
	"time"
)

// ParseQuery 将查询语法解析为elasticsearch的dsl语法
func ParseQuery(expr string) (query bson.M, err error) {
	parseResult, err := parser.ParseExpr(expr)
	if err != nil {
		return
	}
	result := eval(parseResult)
	query = result.(bson.M)
	return
}

// eval 递归解析AST,将表达式解析为elasticsearch的query
func eval(expr ast.Expr) interface{} {
	switch expr := expr.(type) {
	case *ast.BasicLit: // 匹配到数据
		return getlitValue(expr)
	case *ast.BinaryExpr: // 匹配到子树
		x := eval(expr.X)
		y := eval(expr.Y)
		if x == nil || y == nil {
			return errors.New(fmt.Sprintf("%+v, %+v is nil", x, y))
		}
		op := expr.Op
		//fmt.Println(x, op, y)
		//z := fmt.Sprintf("(%s %s %s)", x, op, y)
		//return z
		switch op.String() {
		case "&&":
			return getAndQuery(x, y)
		case "||":
			return getOrQuery(x, y)
		case "!=":
			return getNeQuery(x, y)
		case ">=":
			// 日期范围查询，等同于after
			return getGteQuery(x, y)
		case "<=":
			// 日期范围查询，等同于before
			return getLteQuery(x, y)
		default:
			// 缺省的情况下是==
			return GetEqQuery(x, y)
		}
	case *ast.ParenExpr: // 匹配到括号
		return eval(expr.X)
	case *ast.UnaryExpr: // 匹配到一元表达式
		x := eval(expr.X)
		if x == nil {
			return errors.New(fmt.Sprintf("%+v is nil", x))
		}
		op := expr.Op
		switch op {
		case token.NOT:
			switch x.(type) {
			case bool:
				xb := x.(bool)
				return !xb
			}
		}
		return errors.New(fmt.Sprintf("%x type is not support", expr))
	case *ast.Ident: // 匹配到变量
		return expr.Name
	default:
		return errors.New(fmt.Sprintf("%x type is not support", expr))
	}
}

// getlitValue 获取AST中变量的数据（表达式中的数字为int，转为int64）
func getlitValue(basicLit *ast.BasicLit) interface{} {
	switch basicLit.Kind {
	case token.INT:
		value, err := strconv.ParseInt(basicLit.Value, 10, 64)
		if err != nil {
			return err
		}
		return value
	case token.STRING:
		value, err := strconv.Unquote(basicLit.Value)
		if err != nil {
			return err
		}
		return value
	}
	return errors.New(fmt.Sprintf("%s is not support type", basicLit.Kind))
}

func getAndQuery(x interface{}, y interface{}) bson.M {
	return bson.M{"$and": bson.A{x, y}}
}

func getOrQuery(x interface{}, y interface{}) bson.M {
	return bson.M{"$or": bson.A{x, y}}
}

func GetEqQuery(x interface{}, y interface{}) bson.M {
	key := x.(string)
	switch key {
	// 模糊查询的字段
	case "host", "service", "server", "banner", "title", "header", "body", "cert", "memo":
		return GetRegexQuery(x, y)
	case "app":
		return bson.M{"app": bson.M{"$elemMatch": bson.M{"$regex": y}}}
	case "location":
		ipv4Location := bson.M{"ip.ipv4": bson.M{"$elemMatch": GetRegexQuery("location", y)}}
		ipv6Location := bson.M{"ip.ipv6": bson.M{"$elemMatch": GetRegexQuery("location", y)}}
		return getOrQuery(ipv4Location, ipv6Location)
	case "cdn", "new", "update":
		b, err := strconv.ParseBool(y.(string))
		if err != nil {
			return bson.M{}
		}
		return bson.M{key: b}
	case "ip":
		ip := y.(string)
		// 处理ipv4/ipv4子网/ipv6
		if utils.CheckIPV4(ip) || utils.CheckIPV4Subnet(ip) {
			_ip, _ipNet, err := net.ParseCIDR(y.(string))
			if err != nil {
				return bson.M{"ip.ipv4": bson.M{"$elemMatch": bson.M{"ip": y}}}
			} else {
				ones, bits := _ipNet.Mask.Size()
				_ipStart := utils.IPV4ToUInt32(_ip.String())
				_ipEnd := _ipStart + (1 << (bits - ones)) - 1
				return bson.M{"ip.ipv4": bson.M{"$elemMatch": bson.M{"uint32": bson.M{"$gte": _ipStart, "$lt": _ipEnd}}}}
			}
		}
		if utils.CheckIPV6(ip) {
			//ipv6暂时模糊查询
			return bson.M{"ip.ipv6": bson.M{"$elemMatch": bson.M{"ip": bson.M{"$regex": y}}}}
		}
	case "port":
		ports := strings.Split(y.(string), ",")
		var portsInt []int
		for _, port := range ports {
			p, err := strconv.Atoi(port)
			if err != nil {
				return bson.M{}
			}
			portsInt = append(portsInt, p)
		}
		return bson.M{"port": bson.M{"$in": portsInt}}
	case "create_time", "update_time":
		// 日期查询：查询一天内的记录
		dt, err := time.Parse("2006-01-02", y.(string))
		if err == nil {
			dtNext := time.Unix(dt.Unix(), 0).AddDate(0, 0, 1)
			return bson.M{key: bson.M{"$gte": dt, "$lt": dtNext}}
		} else {
			return bson.M{}
		}
	}
	// 默认是精确匹配的字段,包括port/domain/color/status/icon_hash/category/org等
	return bson.M{x.(string): bson.M{"$eq": y}}
}

func getNeQuery(x interface{}, y interface{}) bson.M {
	key := x.(string)
	switch key {
	// 模糊查询的字段
	case "cdn", "new", "update":
		b, err := strconv.ParseBool(y.(string))
		if err == nil {
			y = b
		}
	}
	return bson.M{x.(string): bson.M{"$ne": y}}
}

func getGteQuery(x interface{}, y interface{}) bson.M {
	if x, y = handleDateSearch(x, y); x == nil || y == nil {
		return bson.M{}
	}
	return bson.M{x.(string): bson.M{"$gte": y}}
}

func getLteQuery(x interface{}, y interface{}) bson.M {
	if x, y = handleDateSearch(x, y); x == nil || y == nil {
		return bson.M{}
	}
	return bson.M{x.(string): bson.M{"$lte": y}}
}

func GetRegexQuery(x interface{}, y interface{}) bson.M {
	return bson.M{x.(string): bson.M{"$regex": y, "$options": "im"}}
}

func handleDateSearch(x interface{}, y interface{}) (interface{}, interface{}) {
	key := x.(string)
	switch key {
	case "create_time", "update_time":
		dt, err := time.Parse("2006-01-02 15:04:05", y.(string))
		if err == nil {
			return x, dt
		}
		dt, err = time.Parse("2006-01-02 15:04", y.(string))
		if err == nil {
			return x, dt
		}
		dt, err = time.Parse("2006-01-02", y.(string))
		if err == nil {
			return x, dt
		}
		return nil, nil
	default:
		return x, y
	}
}
