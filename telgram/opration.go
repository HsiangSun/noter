package telgram

import (
	"errors"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"noter/model"
	"noter/utils/cache"
	"noter/utils/helper"
	"noter/utils/log"
	"noter/utils/orm"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 绑定操作员
func Binding(c tb.Context) error {
	//save to db
	record := model.Admin{}
	record.Gid = fmt.Sprintf("%d", c.Chat().ID)
	record.Uid = c.Sender().Username

	err2 := orm.Gdb.Model(&record).Create(&record).Error
	if err2 != nil {
		log.Sugar.Errorf("save recored error:%s", err2.Error())
		return nil
	}

	//添加到内存
	helper.NoterMap.Store(record.Gid, record.Uid)

	groupName := c.Chat().Title
	userDisplayName := c.Sender().FirstName + c.Sender().LastName
	return c.Reply(fmt.Sprintf("用户:【%s】 已经绑定为:【%s】的记账操作员", userDisplayName, groupName))
}

// 开始记账
func StartNote(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//2.查询汇率
	todayRate, _ := cache.GetTodayRate(c, model.ALL)

	//3.查询手续费
	todayFree, _ := cache.GetTodayFree(c, model.ALL)

	//4.查询交易币种
	currency := cache.GetCurrency(c)

	if currency == "" {
		log.Sugar.Infof("当前群组：%s尚未设置交易币种", c.Chat().Username)
		currency = "USDT"
	}

	inRate := todayRate.InRate
	outRate := todayRate.OutRate

	inFree := todayFree.InFree
	outFree := todayFree.OutFree

	symbal := "```"

	strtmp := `
	|交易币种| %s

	|入款汇率| %.2f
	|出款汇率| %.2f

	|入款费率| %.2f%%
	|出款费率| %.2f%%
	`

	res := fmt.Sprintf(symbal+strtmp+symbal,
		currency, inRate,
		outRate, inFree, outFree,
	)
	//
	txt := helper.EscapeTxt(res)

	return c.Send(txt, tb.ModeMarkdownV2)
}

func SetAdmin(c tb.Context) error {

	err := CheckNoter(c)
	if err != nil {
		return err
	}

	text := c.Text()
	split := strings.Split(text, "@")
	if len(split) != 2 {
		return c.Reply("授权格式错误")
	}

	username := split[1]

	var count int64
	//检查当前用户是否已经授权
	err = orm.Gdb.Model(&model.Admin{}).Where("gid = ? and uid = ? ", fmt.Sprint(c.Chat().ID), username).Count(&count).Error
	if err != nil {
		log.Sugar.Errorf("check gid count error:%s", err.Error)
	}

	if count > 0 {
		return c.Reply(fmt.Sprintf("授权对象:%s 已经被授权过了请勿重复授权"), username)
	}

	//save to db
	record := model.Admin{}
	record.Gid = fmt.Sprintf("%d", c.Chat().ID)
	record.Uid = username

	err2 := orm.Gdb.Model(&record).Create(&record).Error
	if err2 != nil {
		log.Sugar.Errorf("save recored error:%s", err2.Error())
		return nil
	}

	//添加到内存
	helper.AddNorer(record.Gid, record.Uid)
	groupName := c.Chat().Title
	return c.Reply(fmt.Sprintf("用户:【%s】 已经授权为:【%s】的记账操作员", record.Uid, groupName))

}

//TODO 修改完币种之后需要重新设置汇率
func SetCurrency(c tb.Context) error {
	reg := `[A-Z]+`
	regex := regexp.MustCompile(reg)
	currency := regex.FindString(c.Text())

	err2 := orm.Gdb.Model(model.Currency{}).Where("gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Update("currency", currency).Error
	if err2 != nil {
		log.Sugar.Errorf("update currency error:%s", err2.Error())
		return nil
	}
	//cache
	today := time.Now().Format("2006-01-02") + "currency"
	helper.NoterMap.Store(today, currency)

	return c.Reply("币种设置成功")
}

func SetInRate(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//检查今天是否已经设置过入款汇率

	var rate model.Rate

	msg := c.Message().Text
	regex := `[0-9]+\.?[0-9]*`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	err = orm.Gdb.Model(&rate).First(&rate, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			floatRate, err := strconv.ParseFloat(rates, 32)
			if err != nil {
				return c.Reply("指令有误请重试")
			}

			rate.InRate = float32(floatRate)
			rate.Gid = fmt.Sprintf("%d", c.Chat().ID)
			rate.Created = time.Now()
			rate.Operator = c.Sender().Username

			err = orm.Gdb.Model(&rate).Create(&rate).Error
			if err != nil {
				log.Sugar.Errorf("add in_rate to rates db have error:%s", err.Error())
				return c.Reply("设置入款汇率失败")
			}

			return c.Reply(fmt.Sprintf("今日(%s)入款汇率设置为:%s", time.Now().Format("2006-01-02"), rates))

		} else {
			log.Sugar.Errorf("查询当前群组今日入款汇率失败")
		}

	} else {
		floatRate, err := strconv.ParseFloat(rates, 32)
		if err != nil {
			return c.Reply("指令有误请重试")
		}

		rate.InRate = float32(floatRate)

		err = orm.Gdb.Model(&rate).Updates(&rate).Error
		if err != nil {
			log.Sugar.Errorf("add in_rate to rates db have error:%s", err.Error())
			return c.Reply("设置入款汇率失败")
		}

		//更新缓存
		rateKey := fmt.Sprintf("%d%s", c.Chat().ID, "rate")
		helper.NoterMap.Store(rateKey, &rate)

		return c.Reply(fmt.Sprintf("今日(%s)入款汇率设置为:%s", time.Now().Format("2006-01-02"), rates))
	}

	return nil
}

func SetOutRate(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//检查今天是否已经设置过出款汇率

	var rate model.Rate

	msg := c.Message().Text
	regex := `[0-9]+\.?[0-9]*`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	err = orm.Gdb.Model(&rate).First(&rate, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			floatRate, err := strconv.ParseFloat(rates, 32)
			if err != nil {
				return c.Reply("指令有误请重试")
			}

			rate.OutRate = float32(floatRate)
			rate.Gid = fmt.Sprintf("%d", c.Chat().ID)
			rate.Created = time.Now()
			rate.Operator = c.Sender().Username

			err = orm.Gdb.Model(&rate).Create(&rate).Error
			if err != nil {
				log.Sugar.Errorf("add in_rate to rates db have error:%s", err.Error())
				return c.Reply("设置入款汇率失败")
			}

			return c.Reply(fmt.Sprintf("今日(%s)出款汇率设置为:%s", time.Now().Format("2006-01-02"), rates))

		} else {
			log.Sugar.Errorf("查询当前群组今日出款汇率失败")
		}
	} else {
		floatRate, err := strconv.ParseFloat(rates, 32)
		if err != nil {
			return c.Reply("指令有误请重试")
		}

		rate.OutRate = float32(floatRate)

		err = orm.Gdb.Model(&rate).Updates(&rate).Error
		if err != nil {
			log.Sugar.Errorf("update out_rate to rates db have error:%s", err.Error())
			return c.Reply("更新出款汇率失败")
		}
		//更新缓存
		rateKey := fmt.Sprintf("%d%s", c.Chat().ID, "rate")
		helper.NoterMap.Store(rateKey, &rate)
		return c.Reply(fmt.Sprintf("今日(%s)出款汇率设置为:%s", time.Now().Format("2006-01-02"), rates))
	}

	return nil
}

func SetInFree(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//检查今天是否已经设置过入款汇率

	var free model.Free

	msg := c.Message().Text
	regex := `[0-9]+\.?[0-9]*`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	err = orm.Gdb.Model(&free).First(&free, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			floatRate, err := strconv.ParseFloat(rates, 32)
			if err != nil {
				return c.Reply("指令有误请重试")
			}

			free.InFree = float32(floatRate)
			free.Gid = fmt.Sprintf("%d", c.Chat().ID)
			free.Created = time.Now()
			free.Operator = c.Sender().Username

			err = orm.Gdb.Model(&free).Create(&free).Error
			if err != nil {
				log.Sugar.Errorf("add in_rate to rates db have error:%s", err.Error())
				return c.Reply("设置入款费率失败")
			}

			return c.Reply(fmt.Sprintf("今日(%s)入款费率设置为:%s", time.Now().Format("2006-01-02"), rates))

		} else {
			log.Sugar.Errorf("查询当前群组今日入款费率失败")
		}

	} else {
		floatRate, err := strconv.ParseFloat(rates, 32)
		if err != nil {
			return c.Reply("指令有误请重试")
		}

		free.InFree = float32(floatRate)

		err = orm.Gdb.Model(&free).Updates(&free).Error
		if err != nil {
			log.Sugar.Errorf("add in_free to rates db have error:%s", err.Error())
			return c.Reply("设置入款费率失败")
		}

		//更新缓存
		freeKey := fmt.Sprintf("%d%s", c.Chat().ID, "free")
		helper.NoterMap.Store(freeKey, &free)

		return c.Reply(fmt.Sprintf("今日(%s)入款费率设置为:%s", time.Now().Format("2006-01-02"), rates))
	}

	return nil

}

func SetOutFree(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//检查今天是否已经设置过入款汇率

	var free model.Free

	msg := c.Message().Text
	regex := `[0-9]+\.?[0-9]*`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	err = orm.Gdb.Model(&free).First(&free, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			floatRate, err := strconv.ParseFloat(rates, 32)
			if err != nil {
				return c.Reply("指令有误请重试")
			}

			free.OutFree = float32(floatRate)
			free.Gid = fmt.Sprintf("%d", c.Chat().ID)
			free.Created = time.Now()
			free.Operator = c.Sender().Username

			err = orm.Gdb.Model(&free).Create(&free).Error
			if err != nil {
				log.Sugar.Errorf("add out_free to rates db have error:%s", err.Error())
				return c.Reply("设置入款费率失败")
			}

			return c.Reply(fmt.Sprintf("今日(%s)出款费率设置为:%s", time.Now().Format("2006-01-02"), rates))

		} else {
			log.Sugar.Errorf("查询当前群组今日出款费率失败")
		}

	} else {
		floatRate, err := strconv.ParseFloat(rates, 32)
		if err != nil {
			return c.Reply("指令有误请重试")
		}

		free.OutFree = float32(floatRate)

		err = orm.Gdb.Model(&free).Updates(&free).Error
		if err != nil {
			log.Sugar.Errorf("add in_free to rates db have error:%s", err.Error())
			return c.Reply("设置入款费率失败")
		}

		//更新缓存
		freeKey := fmt.Sprintf("%d%s", c.Chat().ID, "free")
		helper.NoterMap.Store(freeKey, &free)

		return c.Reply(fmt.Sprintf("今日(%s)出款费率设置为:%s", time.Now().Format("2006-01-02"), rates))
	}

	return nil

}

//任何人都可以显示账单
func ShowBill(c tb.Context) error {
	//查询今日费率
	rate, err := cache.GetTodayRate(c, model.ALL)
	if rate == nil && err != nil {
		return c.Reply(err.Error())
	}
	//查询今日手续费
	free, err := cache.GetTodayFree(c, model.ALL)
	if free == nil && err != nil {
		return c.Reply(err.Error())
	}

	//获取当前币种设置
	currency := cache.GetCurrency(c)
	if currency == "" {
		currency = "USDT"
	}

	//查询今日所有交易订单
	var bills []model.Bill
	err = orm.Gdb.Model(&model.Bill{}).Find(&bills, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		log.Sugar.Errorf("get blance have error:%s", err.Error())
	}

	billInCount := 0
	billOutCount := 0
	var billInAmount float64 = 0
	var billOutAmount float64 = 0

	billTmp := "%s    %.2f \n"
	billInStr := ""
	billOutStr := ""

	for _, bi := range bills {
		if bi.Direction == model.PAY_IN {
			billInCount++
			billInAmount += bi.Amount
			billInStr += fmt.Sprintf(billTmp, bi.Created.Format("2006-01-02 15:04:05"), bi.Amount)
		}
		if bi.Direction == model.PAY_OUT {
			billOutCount++
			billOutAmount += bi.Amount
			billOutStr += fmt.Sprintf(billTmp, bi.Created.Format("2006-01-02 15:04:05"), bi.Amount)
		}
	}

	var totalIn float64 = 0
	var inTax float64 = 0
	var billInTotal float64 = 0
	if rate.InRate != 0 && free.InFree != 0 {
		//计算入款手续费
		inTax = billInAmount * float64(free.InFree) / 100
		//入款总计
		billInTotal = billInAmount - inTax
		//根据汇率算成对应单位
		totalIn = billInTotal / float64(rate.InRate)
	}

	var outTax float64 = 0
	var billOutTotal float64 = 0
	var totalOut float64 = 0
	if rate.OutRate != 0 && free.OutFree != 0 {
		//计算出款手续费
		outTax = billOutAmount * float64(free.OutFree) / 100
		//出款总计
		billOutTotal = billOutAmount + outTax
		//根据出汇率算成对应单位
		totalOut = billOutTotal / float64(rate.OutRate)
	}

	rspStr := `
 入款(%d笔):
	%s
出款(%d笔):
	%s

入款费率：%.2f%%
入款汇率：%.2f
入款总数：%.2f
入款总计：%.2f｜%.2f %s

出款费率：%.2f%%
出款汇率：%.2f
出款总数：%.2f
出款总计：%.2f｜%.2f %s
`

	txt := fmt.Sprintf(rspStr,
		billInCount,
		billInStr,

		billOutCount,
		billOutStr,

		free.InFree, rate.InRate,
		billInAmount,
		billInTotal, totalIn, currency,

		free.OutFree, rate.OutRate,
		billOutAmount,
		billOutTotal, totalOut, currency,
	)

	return c.Send(txt)
}

// 记账
func Record(c tb.Context, direction int8) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//查询今日汇率
	rate, err := cache.GetTodayRate(c, direction)
	if err != nil {
		if direction == model.PAY_IN && strings.Contains(err.Error(), "尚未设置入款汇率") {
			return c.Reply(err.Error())
		}
		if direction == model.PAY_OUT && strings.Contains(err.Error(), "尚未设置出款汇率") {
			return c.Reply(err.Error())
		}
	}

	//查询今日手续费
	free, err := cache.GetTodayFree(c, direction)
	if err != nil {
		if direction == model.PAY_IN && strings.Contains(err.Error(), "尚未设置入款费率") {
			return c.Reply(err.Error())
		}
		if direction == model.PAY_OUT && strings.Contains(err.Error(), "尚未设置出款费率") {
			return c.Reply(err.Error())
		}
	}

	//获取当前币种设置
	currency := cache.GetCurrency(c)
	if currency == "" {
		currency = "USDT"
	}

	order, amount, err := helper.GetOrderAndAmount(c.Text())
	if err != nil {
		log.Sugar.Errorf("%s 无法记录该笔信息", c.Text())
		return nil
	}

	bill := model.Bill{}
	bill.Gid = fmt.Sprintf("%d", c.Chat().ID)
	bill.Created = time.Now()
	bill.Order = order
	bill.Amount = amount
	bill.Direction = direction
	bill.Operator = c.Sender().Username

	err = orm.Gdb.Create(&bill).Error
	if err != nil {
		log.Sugar.Errorf("insert to bills have error:%s", err.Error())
		return err
	}

	//查询今日所有交易订单
	var bills []model.Bill
	err = orm.Gdb.Model(&model.Bill{}).Find(&bills, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		log.Sugar.Errorf("get blance have error:%s", err.Error())
	}

	billInCount := 0
	billOutCount := 0
	var billInAmount float64 = 0
	var billOutAmount float64 = 0

	billTmp := "%s    %.2f \n"
	billInStr := ""
	billOutStr := ""

	for _, bi := range bills {
		if bi.Direction == model.PAY_IN {
			billInCount++
			billInAmount += bi.Amount
			billInStr += fmt.Sprintf(billTmp, bi.Created.Format("2006-01-02 15:04:05"), bi.Amount)
		}
		if bi.Direction == model.PAY_OUT {
			billOutCount++
			billOutAmount += bi.Amount
			billOutStr += fmt.Sprintf(billTmp, bi.Created.Format("2006-01-02 15:04:05"), bi.Amount)
		}
	}

	var totalIn float64 = 0
	var inTax float64 = 0
	var billInTotal float64 = 0
	if rate.InRate != 0 && free.InFree != 0 {
		//计算入款手续费
		inTax = billInAmount * float64(free.InFree) / 100
		//入款总计
		billInTotal = billInAmount - inTax
		//根据汇率算成对应单位
		totalIn = billInTotal / float64(rate.InRate)
	}

	var outTax float64 = 0
	var billOutTotal float64 = 0
	var totalOut float64 = 0
	if rate.OutRate != 0 && free.OutFree != 0 {
		//计算出款手续费
		outTax = billOutAmount * float64(free.OutFree) / 100
		//出款总计
		billOutTotal = billOutAmount + outTax
		//根据出汇率算成对应单位
		totalOut = billOutTotal / float64(rate.OutRate)
	}

	//入款进度
	//inProcess := totalIn * 100 / balanceInAmount

	rspStr := `
 入款(%d笔):
	%s
出款(%d笔):
	%s

入款费率：%.2f%%
入款汇率：%.2f
入款总数：%.2f
入款总计：%.2f｜%.2f %s

出款费率：%.2f%%
出款汇率：%.2f
出款总数：%.2f
出款总计：%.2f｜%.2f %s
`

	txt := fmt.Sprintf(rspStr,
		billInCount,
		billInStr,

		billOutCount,
		billOutStr,

		free.InFree, rate.InRate,
		billInAmount,
		billInTotal, totalIn, currency,

		free.OutFree, rate.OutRate,
		billOutAmount,
		billOutTotal, totalOut, currency,
	)

	return c.Send(txt)

}

//清账
func clean(c tb.Context) error {

	//清除汇率
	rateKey := fmt.Sprintf("%d%s", c.Chat().ID, "rate")
	helper.NoterMap.Delete(rateKey)
	err := orm.Gdb.Model(&model.Rate{}).Where("gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Delete(&model.Rate{}).Error
	if err != nil {
		log.Sugar.Errorf("delete rates have error:%s", err.Error())
	}

	//清除费率
	freeKey := fmt.Sprintf("%d%s", c.Chat().ID, "free")
	helper.NoterMap.Delete(freeKey)
	err = orm.Gdb.Model(&model.Free{}).Where("gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Delete(&model.Free{}).Error
	if err != nil {
		log.Sugar.Errorf("delete frees have error:%s", err.Error())
	}

	//删除账单
	err = orm.Gdb.Model(&model.Bill{}).Where("gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Delete(&model.Bill{}).Error
	if err != nil {
		log.Sugar.Errorf("delete bills have error:%s", err.Error())
	}
	//
	//wg := sync.WaitGroup{}
	//wg.Add(3)
	//
	//go func() {
	//	defer wg.Done()
	//	//删除账单
	//	err := orm.Gdb.Model(&model.Bill{}).Where("gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Delete(&model.Bill{}).Error
	//	if err != nil {
	//		log.Sugar.Errorf("delete bills have error:%s", err.Error())
	//	}
	//}()
	//
	//go func() {
	//	defer wg.Done()
	//	//清除汇率
	//	rateKey := fmt.Sprintf("%d%s", c.Chat().ID, "rate")
	//	helper.NoterMap.Delete(rateKey)
	//	err := orm.Gdb.Model(&model.Rate{}).Where("gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Delete(&model.Rate{}).Error
	//	if err != nil {
	//		log.Sugar.Errorf("delete rates have error:%s", err.Error())
	//	}
	//}()
	//
	//go func() {
	//	defer wg.Done()
	//	//清除费率
	//	freeKey := fmt.Sprintf("%d%s", c.Chat().ID, "free")
	//	helper.NoterMap.Delete(freeKey)
	//	err := orm.Gdb.Model(&model.Free{}).Where("gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Delete(&model.Free{}).Error
	//	if err != nil {
	//		log.Sugar.Errorf("delete frees have error:%s", err.Error())
	//	}
	//}()
	//wg.Wait()
	return c.Reply(fmt.Sprintf("今日(%s)清账完成", time.Now().Format("2006-01-02")))
}

//机器人帮助
func Help(c tb.Context) error {

	hrlpMsg := `
1.如何设置汇率？
设置入(出)款汇率xxx.xxxx【数字最大保留小数点后4位】
2.如何设置费率？
设置入(出)款费率xxx.xxxx【数字最大保留小数点后4位】
3.如何查看设置项？
记账【查看当前费率和汇率设置以及币种】
4.如何查看账单
查账
5.如何设置币种？
设置币种XXX【大写字母 例如CNY、USDT】
6.如何记账？
（1.）如果有账单号:xxxx+金额
（2.）如果没有账单号:+金额
（3.）其中+为代收 -为代付
7.如何清账？
清账
8.如何授权
授权请联系[@HsiangSun](https://t.me/HsiangSun)
`

	return c.Send(hrlpMsg)
}
