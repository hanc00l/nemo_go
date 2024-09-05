package es

import (
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

// forked from  https://github.com/wangxin1248/gparse

// ParseQuery 将查询语法解析为elasticsearch的dsl语法
func ParseQuery(expr string) (query types.Query, err error) {
	parseResult, err := parser.ParseExpr(expr)
	if err != nil {
		return
	}
	result := eval(parseResult)
	query = result.(types.Query)
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
			return getNotQuery(x, y)
		case ">=":
			// 日期范围查询，等同于after
			return getRangeGteQuery(x, y)
		case "<=":
			// 日期范围查询，等同于before
			return getRangeLteQuery(x, y)
		default:
			// 缺省的情况下是==
			return getEqualQuery(x, y)
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

func getAndQuery(x, y interface{}) types.Query {
	q := types.Query{
		Bool: &types.BoolQuery{
			Must: []types.Query{
				x.(types.Query), y.(types.Query),
			}},
	}
	return q
}

func getOrQuery(x, y interface{}) types.Query {
	q := types.Query{
		Bool: &types.BoolQuery{
			Should: []types.Query{
				x.(types.Query), y.(types.Query),
			}},
	}
	return q
}
func getNotQuery(x, y interface{}) types.Query {
	q := types.Query{
		Bool: &types.BoolQuery{
			MustNot: []types.Query{
				{
					MatchPhrase: map[string]types.MatchPhraseQuery{
						x.(string): {Query: y.(string)},
					},
				},
			},
		},
	}
	return q
}

func getRangeGteQuery(x, y interface{}) types.Query {
	dateQuery := y.(string)
	//默认的日期格式是yyyy-MM-dd
	q := types.Query{
		Range: map[string]types.RangeQuery{
			x.(string): types.DateRangeQuery{
				Gte: &dateQuery,
			},
		},
	}
	return q
}

func getRangeLteQuery(x, y interface{}) types.Query {
	dateQuery := y.(string)
	//默认的日期格式是yyyy-MM-dd
	q := types.Query{
		Range: map[string]types.RangeQuery{
			x.(string): types.DateRangeQuery{
				Lte: &dateQuery,
			},
		},
	}
	return q
}

func getEqualQuery(x, y interface{}) types.Query {
	q := types.Query{MatchPhrase: map[string]types.MatchPhraseQuery{
		x.(string): {
			Query: y.(string),
		}},
	}
	return q
}
