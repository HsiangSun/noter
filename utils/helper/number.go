package helper

import (
	"regexp"
	"strconv"
)

func GetNumberFromString(str string) float64 {
	regex := `[0-9]+.[0-9]+`
	mustCompile := regexp.MustCompile(regex)
	res := mustCompile.FindString(str)

	number, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return 0
	}
	return number
}

func GetOrderAndAmount(str string) (order string, amount float64, err error) {
	regex := `^(\w+)\s*(\-|\+)+\s*([0-9]+.?[0-9+]+)$`
	mustCompile := regexp.MustCompile(regex)
	matchs := mustCompile.FindStringSubmatch(str)

	//for i, i2 := range matchs {
	//	fmt.Printf("%d ==> %s \n", i, i2)
	//}

	order = matchs[1]
	amount, err = strconv.ParseFloat(matchs[3], 64)
	if err != nil {
		return
	}
	return
}
