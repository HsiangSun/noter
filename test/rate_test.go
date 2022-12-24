package test

import (
	"fmt"
	"noter/utils/helper"
	"regexp"
	"strings"
	"testing"
)

func TestRate(t *testing.T) {
	msg := "设置入款汇率22.33 45.12"

	contains := strings.Contains(msg, "置入款汇率")
	if contains {
		fmt.Println("可以设置汇率")
	}

	regex := `[0-9]+.[0-9]+`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	fmt.Println(rates)

}

func TestNumberFromString(t *testing.T) {
	str := "申请入款12.58"
	num := helper.GetNumberFromString(str)
	fmt.Println(num)
}

func TestRecord(t *testing.T) {

	str := "abc+100"

	regex := `(.*)?\s*(\+|\-)\s*(\d+.\d+)`
	mustCompile := regexp.MustCompile(regex)
	matchs := mustCompile.FindStringSubmatch(str)

	fmt.Println(matchs)
}

func TestOrderAndAmount(t *testing.T) {
	str := `20221217173105_172_1927_OC +   1200.00`
	order, a, err := helper.GetOrderAndAmount(str)
	if err != nil {
		t.Log(err)
	}
	fmt.Println(order)
	fmt.Println(a)
}

func TestCurrency(t *testing.T) {
	reg := `[A-Z]+`
	regex := regexp.MustCompile(reg)

	findString := regex.FindString("设置币种为USDT")

	fmt.Println(findString)

}

func TestTableSysbal(t *testing.T) {

	//| Symbol | Price | Change |
	//|--------|-------|--------|
	//| ABC    | 20.85 |  1.626 |
	//| DEF    | 78.95 |  0.099 |
	//| GHI    | 23.45 |  0.192 |
	//| JKL    | 98.85 |  0.292 |

	str1 := `` + "`" + "`" + "`" + ``

	center := `| Symbol2 | Price2 | Change |
	|--------|-------|--------|
	| ABC    | 20.85 |  1.626 |
	| DEF    | 78.95 |  0.099 |
	| GHI    | 23.45 |  0.192 |
	| JKL    | 98.85 |  0.292 |`

	str2 := `` + "`" + "`" + "`" + ``

	fmt.Println(str1 + center + str2)

}
