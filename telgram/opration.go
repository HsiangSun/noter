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
	"time"
)

// 绑定操作员
func Binding(c tb.Context) error {
	var count int64
	//检查当前群组是否已经绑定过了
	err := orm.Gdb.Model(&model.Admin{}).
		Where("gid = ?", fmt.Sprint(c.Chat().ID)).Count(&count)
	if err != nil {
		log.Sugar.Errorf("check gid count error:%s", err.Error)
	}

	if count > 0 {
		log.Sugar.Infof("用户:%s 试图重新往群:%s 绑定记录员", c.Sender().Username, c.Chat().Title)
		return c.Delete()
	}

	//save to db
	record := model.Admin{}
	record.Gid = fmt.Sprintf("%d", c.Chat().ID)
	record.Uid = fmt.Sprintf("%d", c.Sender().ID)
	record.Currency = "USDT"

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

	//查询所有的设置项
	//1.查询余额
	var balances []model.Balance
	err = orm.Gdb.Model(&model.Balance{}).Find(&balances, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error

	var balanceInAmount float64 = 0
	var balanceOutAmount float64 = 0

	for _, balance := range balances {
		if balance.PayIn != 0 {
			balanceInAmount += balance.PayIn
		}
		if balance.PayOut != 0 {
			balanceOutAmount += balance.PayOut
		}
	}

	//2.查询汇率
	var rate model.Rate
	err = orm.Gdb.Model(&model.Rate{}).Find(&rate, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error

	//3.查询手续费
	var free model.Free
	err = orm.Gdb.Model(&model.Free{}).Find(&free, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error

	//4.查询交易币种
	currency := cache.GetCurrency(c)

	inRate := rate.InRate
	outRate := rate.OutRate

	inFree := free.InFree
	outFree := free.OutFree

	symbal := "```"

	strtmp := `
	|交易币种| %s

	|入款金额| %.2f
	|出款金额| %.2f

	|入款汇率| %.2f
	|出款汇率| %.2f

	|入款费率| %.2f%%
	|出款费率| %.2f%%
	`

	res := fmt.Sprintf(symbal+strtmp+symbal,
		currency,
		balanceInAmount,
		balanceOutAmount, inRate,
		outRate, inFree, outFree,
	)
	//
	txt := helper.EscapeTxt(res)

	return c.Send(txt, tb.ModeMarkdownV2)
}

//TODO 修改完币种之后需要重新设置汇率
func SetCurrency(c tb.Context) error {
	reg := `[A-Z]+`
	regex := regexp.MustCompile(reg)
	currency := regex.FindString(c.Text())

	fmt.Println(currency)

	err2 := orm.Gdb.Model(model.Admin{}).Where("gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Update("currency", currency).Error
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
	regex := `[0-9]+.[0-9]+`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	err = orm.Gdb.Model(&rate).First(&rate, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
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

	}

	if rate.InRate == 0 {

		floatRate, err := strconv.ParseFloat(rates, 32)
		if err != nil {
			return c.Reply("指令有误请重试")
		}

		rate.InRate = float32(floatRate)

		err = orm.Gdb.Model(&rate).Create(&rate).Error
		if err != nil {
			log.Sugar.Errorf("add in_rate to rates db have error:%s", err.Error())
			return c.Reply("设置入款汇率失败")
		}

		return c.Reply(fmt.Sprintf("今日(%s)入款汇率设置为:%s", time.Now().Format("2006-01-02"), rates))

	}

	return c.Reply(fmt.Sprintf("⚠今日入款汇率已经设置为:%.2f", rate.InRate))

}

func SetOutRate(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//检查今天是否已经设置过出款汇率

	var rate model.Rate

	msg := c.Message().Text
	regex := `[0-9]+.[0-9]+`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	err = orm.Gdb.Model(&rate).First(&rate, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
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
	}

	if rate.OutRate == 0 {

		floatRate, err := strconv.ParseFloat(rates, 32)
		if err != nil {
			return c.Reply("指令有误请重试")
		}

		rate.OutRate = float32(floatRate)

		err = orm.Gdb.Model(&rate).Updates(&rate).Error
		if err != nil {
			log.Sugar.Errorf("add in_rate to rates db have error:%s", err.Error())
			return c.Reply("设置入款汇率失败")
		}

		return c.Reply(fmt.Sprintf("今日(%s)出款汇率设置为:%s", time.Now().Format("2006-01-02"), rates))
	}

	return c.Reply(fmt.Sprintf("⚠今日出款汇率已经设置为:%.2f", rate.OutRate))
}

func SetInFree(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//检查今天是否已经设置过入款汇率

	var free model.Free

	msg := c.Message().Text
	regex := `[0-9]+.[0-9]+`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	err = orm.Gdb.Model(&free).First(&free, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
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

	}

	if free.InFree == 0 {

		floatRate, err := strconv.ParseFloat(rates, 32)
		if err != nil {
			return c.Reply("指令有误请重试")
		}

		free.InFree = float32(floatRate)

		err = orm.Gdb.Model(&free).Create(&free).Error
		if err != nil {
			log.Sugar.Errorf("add in_free to rates db have error:%s", err.Error())
			return c.Reply("设置入款费率失败")
		}

		return c.Reply(fmt.Sprintf("今日(%s)入款费率设置为:%s", time.Now().Format("2006-01-02"), rates))

	}

	return c.Reply(fmt.Sprintf("⚠今日入款费率已经设置为:%.2f", free.InFree))

}

func SetOutFree(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	//检查今天是否已经设置过入款汇率

	var free model.Free

	msg := c.Message().Text
	regex := `[0-9]+.[0-9]+`
	mustCompile := regexp.MustCompile(regex)
	rates := mustCompile.FindString(msg)

	err = orm.Gdb.Model(&free).First(&free, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
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

	}

	if free.OutFree == 0 {

		floatRate, err := strconv.ParseFloat(rates, 32)
		if err != nil {
			return c.Reply("指令有误请重试")
		}

		free.OutFree = float32(floatRate)

		err = orm.Gdb.Model(&free).Update("out_free", free.OutFree).Error
		if err != nil {
			log.Sugar.Errorf("add in_free to rates db have error:%s", err.Error())
			return c.Reply("设置入款费率失败")
		}

		return c.Reply(fmt.Sprintf("今日(%s)出款费率设置为:%s", time.Now().Format("2006-01-02"), rates))

	}

	return c.Reply(fmt.Sprintf("⚠今日入款费率已经设置为:%.2f", free.InFree))

}

// 设置入款
func PayIn(c tb.Context) error {

	err := CheckNoter(c)
	if err != nil {
		return err
	}

	num := helper.GetNumberFromString(c.Text())

	payin := model.Balance{}
	payin.Gid = fmt.Sprintf("%d", c.Chat().ID)
	payin.PayIn = num
	payin.PayOut = 0
	payin.Created = time.Now()

	err = orm.Gdb.Create(&payin).Error
	if err != nil {
		log.Sugar.Errorf("insert balance error:%s", err.Error())
	}

	return c.Reply("设置入款金额成功")
}

// 设置出款
func PayOut(c tb.Context) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	num := helper.GetNumberFromString(c.Text())

	var dbBalance model.Balance
	err = orm.Gdb.Model(&model.Balance{}).First(&dbBalance, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			//没有设置金额信息
			return c.Reply("尚未设置入款信息,请先设置入款信息")
		}

		log.Sugar.Errorf("select  balce error:%s", err.Error())
	}

	//出款余额不足
	if num > dbBalance.PayIn {
		return c.Reply("出款余额不足")
	}

	//payOut := model.Balance{}
	//payOut.Gid = fmt.Sprintf("%d", c.Chat().ID)
	//payOut.PayOut = num
	//payOut.Created = time.Now()

	err = orm.Gdb.Model(model.Balance{}).Update("pay_out", num).Error
	if err != nil {
		log.Sugar.Errorf("update balance pay_out error:%s", err.Error())
	}

	return c.Reply(fmt.Sprintf("出款成功%.2f", num))
}

// 记账
func Record(c tb.Context, direction int8) error {
	err := CheckNoter(c)
	if err != nil {
		return err
	}

	order, amount, err := helper.GetOrderAndAmount(c.Text())
	if err != nil {
		log.Sugar.Errorf("%s 无法记录该笔信息", c.Text())
		return err
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

	//查询出入金
	//var balances []model.Balance
	//err = orm.Gdb.Model(&model.Balance{}).Find(&balances, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	//if err != nil {
	//	log.Sugar.Errorf("get blance have error:%s", err.Error())
	//}

	//if len(balances) == 0 {
	//	return c.Send("请先设置入款与出款金额")
	//}
	//查询今日费率
	rate, err := cache.GetTodayRate(c)
	if rate == nil && err != nil {
		return c.Reply(err.Error())
	}
	//查询今日手续费
	free, err := cache.GetTodayFree(c)
	if free == nil && err != nil {
		return c.Reply(err.Error())
	}
	//查询今日所有交易订单
	var bills []model.Bill
	err = orm.Gdb.Model(&model.Bill{}).Find(&bills, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		log.Sugar.Errorf("get blance have error:%s", err.Error())
	}
	//获取当前币种设置
	currency := cache.GetCurrency(c)

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

	//计算入款手续费
	inTax := billInAmount * float64(free.InFree) / 100
	//入款总计
	billInTotal := billInAmount - inTax
	//根据汇率算成对应单位
	totalIn := billInTotal / float64(rate.InRate)

	//计算出款手续费
	outTax := billOutAmount * float64(free.OutFree) / 100
	//出款总计
	billOutTotal := billOutAmount + outTax
	//根据出汇率算成对应单位
	totalOut := billOutTotal / float64(rate.OutRate)

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
		billInCount, billInStr,
		billOutCount, billOutStr,

		free.InFree, rate.InRate,
		billInAmount,
		billInTotal, totalIn, currency,

		free.OutFree, rate.OutRate,
		billOutAmount,
		billOutTotal, totalOut, currency,
	)

	return c.Send(txt)

}
