package helper

import (
	"errors"
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
	//regex := `(.*)?\s*(\+|\-)\s*(\d+.\d+)`
	regex := `(.*)(\+|\-)\s?([0-9]+\.?[0-9]*)`
	mustCompile := regexp.MustCompile(regex)
	matchs := mustCompile.FindStringSubmatch(str)

	//for i, i2 := range matchs {
	//	fmt.Printf("%d ==> %s \n", i, i2)
	//}

	if len(matchs) == 0 {
		return "", 0, errors.New("无法匹配")
	}

	if len(matchs) == 4 {
		order = matchs[1]

	} else {
		order = ""
	}

	amount, err = strconv.ParseFloat(matchs[len(matchs)-1], 64)
	if err != nil {
		return
	}

	return
}
