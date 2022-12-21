package telgram

import (
	"errors"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"noter/model"
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

	err2 := orm.Gdb.Model(&record).Create(&record).Error
	if err2 != nil {
		log.Sugar.Errorf("save recored error:%s", err2.Error())
		return nil
	}

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
	return c.Send("智能记账助手开始记账")
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
	var balances []model.Balance
	err = orm.Gdb.Model(&model.Balance{}).Find(&balances, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		log.Sugar.Errorf("get blance have error:%s", err.Error())
	}

	balanceInCount := 0
	balanceOutCount := 0
	var balanceInAmount float64 = 0
	var balanceOutAmount float64 = 0

	balanceTmp := `
%s    %.2f
`
	balanceInStr := ""
	balanceOutStr := ""

	for _, ba := range balances {
		if ba.PayIn != 0 {
			balanceInCount += 1
			balanceInAmount += ba.PayIn
			balanceInStr += fmt.Sprintf(balanceTmp, ba.Created.Format("2006-01-02 15:04:05"), ba.PayIn)
		}
		if ba.PayOut != 0 {
			balanceOutCount += 1
			balanceOutAmount += ba.PayOut
			balanceOutStr += fmt.Sprintf(balanceTmp, ba.Created.Format("2006-01-02 15:04:05"), ba.PayOut)
		}

	}

	var billes = []model.Bill{}
	err = orm.Gdb.Model(&model.Bill{}).Find(&billes, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
	if err != nil {
		log.Sugar.Errorf("get bills have error:%s", err.Error())
	}

	//统计次数
	inCount := 0
	outCount := 0
	for _, bi := range billes {
		if bi.Direction == model.PAY_IN {
			inCount++
		}
		if bi.Direction == model.PAY_OUT {
			outCount++
		}
	}

	fmt.Println(billes)

	rspStr := `
		入款（1笔）：
		  %s
		
		下发（2笔）：
		  %s
`

	txt := fmt.Sprintf(rspStr, balanceInStr, balanceOutStr)

	return c.Send(txt)

}
