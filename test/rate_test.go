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

func TestOrderAndAmount(t *testing.T) {
	str := `20221217173105_172_1927_OC +   1200.00`
	order, a, err := helper.GetOrderAndAmount(str)
	if err != nil {
		t.Log(err)
	}
	fmt.Println(order)
	fmt.Println(a)
}
