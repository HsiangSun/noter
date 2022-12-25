package cache

import (
	"errors"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"noter/model"
	"noter/utils/helper"
	"noter/utils/log"
	"noter/utils/orm"
	"time"
)

//获取当前的币种设置
func GetCurrency(c tb.Context) string {
	today := time.Now().Format("2006-01-02") + "currency"

	res, ok := helper.NoterMap.Load(fmt.Sprintf("%s", today))
	if !ok {
		var admin model.Currency
		err := orm.Gdb.Model(&model.Currency{}).Find(&admin, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
		if err != nil {
			log.Sugar.Errorf("get rate have error:%s", err.Error())
			return ""
		}
		//缓存起来
		helper.NoterMap.Store(today, admin.Currency)
		return admin.Currency
	}
	return res.(string)
}

//获取今天的汇率
func GetTodayRate(c tb.Context) (*model.Rate, error) {

	today := time.Now().Format("2006-01-02") + "rate"

	res, ok := helper.NoterMap.Load(fmt.Sprintf("%s", today))
	if !ok {
		//去数据库查找
		var rate model.Rate
		err := orm.Gdb.Model(&model.Rate{}).Find(&rate, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
		if err != nil {
			log.Sugar.Errorf("get rate have error:%s", err.Error())
			return nil, err
		}
		if rate.ID == 0 {
			return nil, errors.New("尚未设置汇率")
		}
		if rate.InRate == 0 {
			return nil, errors.New("尚未入款汇率")
		}
		if rate.OutRate == 0 {
			return nil, errors.New("尚未出款汇率")
		}
		//缓存起来
		helper.NoterMap.Store(today, &rate)
		return &rate, nil
	}

	return res.(*model.Rate), nil
}

//获取今天的交易手续费
func GetTodayFree(c tb.Context) (*model.Free, error) {

	today := time.Now().Format("2006-01-02") + "free"

	res, ok := helper.NoterMap.Load(fmt.Sprintf("%s", today))
	if !ok {
		//去数据库查找
		var free model.Free
		err := orm.Gdb.Model(&model.Free{}).Find(&free, "created >= date('now','start of day') and gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
		if err != nil {
			log.Sugar.Errorf("get free have error:%s", err.Error())
			return nil, err
		}
		if free.ID == 0 {
			return nil, errors.New("尚未设置手续费")
		}
		if free.InFree == 0 {
			return nil, errors.New("尚未设置入款费率")
		}
		if free.OutFree == 0 {
			return nil, errors.New("尚未设置出款费率")
		}
		//缓存起来
		helper.NoterMap.Store(today, &free)
		return &free, nil
	}

	return res.(*model.Free), nil
}
