package cache

import (
	"errors"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"noter/model"
	"noter/utils/helper"
	"noter/utils/log"
	"noter/utils/orm"
)

//获取当前群设置的币种
func GetCurrency(c tb.Context) string {
	currencyKey := fmt.Sprintf("%d%s", c.Chat().ID, "currency")

	res, ok := helper.NoterMap.Load(fmt.Sprintf("%s", currencyKey))
	if !ok {
		var currency model.Currency
		err := orm.Gdb.Model(&model.Currency{}).Find(&currency, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
		if err != nil {
			log.Sugar.Errorf("get rate have error:%s", err.Error())
			return ""
		}
		//缓存起来
		helper.NoterMap.Store(currencyKey, currency.Currency)
		return currency.Currency
	}
	return res.(string)
}

//获取当前群设置的汇率
func GetTodayRate(c tb.Context, direction int8) (*model.Rate, error) {

	rateKey := fmt.Sprintf("%d%s", c.Chat().ID, "rate")

	res, ok := helper.NoterMap.Load(fmt.Sprintf("%s", rateKey))
	if !ok {
		//去数据库查找
		var rate model.Rate
		err := orm.Gdb.Model(&model.Rate{}).Find(&rate, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
		if err != nil {
			log.Sugar.Errorf("get rate have error:%s", err.Error())
			return nil, err
		}
		if direction == model.PAY_IN && rate.InRate == 0 {
			return &rate, errors.New("尚未设置入款汇率")
		}
		if direction == model.PAY_OUT && rate.OutRate == 0 {
			return &rate, errors.New("尚未设置出款汇率")
		}

		if direction == model.ALL {
			if rate.InRate == 0 && rate.OutRate == 0 {
				return &rate, errors.New("尚未设置汇率")
			}
			if rate.InRate == 0 {
				return &rate, errors.New("尚未设置入款汇率")
			}
			if rate.OutRate == 0 {
				return &rate, errors.New("尚未设置入款汇率")
			}
		}

		//缓存起来
		helper.NoterMap.Store(rateKey, &rate)
		return &rate, nil
	}

	rate := res.(*model.Rate)

	if direction == model.PAY_IN && rate.InRate == 0 {
		return rate, errors.New("尚未设置入款汇率")
	}
	if direction == model.PAY_OUT && rate.OutRate == 0 {
		return rate, errors.New("尚未设置出款汇率")
	}

	if direction == model.ALL {
		if rate.InRate == 0 && rate.OutRate == 0 {
			return rate, errors.New("尚未设置汇率")
		}
		if rate.InRate == 0 {
			return rate, errors.New("尚未设置入款汇率")
		}
		if rate.OutRate == 0 {
			return rate, errors.New("尚未设置入款汇率")
		}
	}

	return res.(*model.Rate), nil
}

//获取当前群设置的交易手续费
func GetTodayFree(c tb.Context, direction int8) (*model.Free, error) {

	freeKey := fmt.Sprintf("%d%s", c.Chat().ID, "free")

	res, ok := helper.NoterMap.Load(fmt.Sprintf("%s", freeKey))
	if !ok {
		//去数据库查找
		var free model.Free
		err := orm.Gdb.Model(&model.Free{}).Find(&free, "gid = ?", fmt.Sprintf("%d", c.Chat().ID)).Error
		if err != nil {
			log.Sugar.Errorf("get free have error:%s", err.Error())
			return nil, err
		}

		if direction == model.PAY_IN && free.InFree == 0 {
			return &free, errors.New("尚未设置入款费率")
		}
		if direction == model.PAY_OUT && free.OutFree == 0 {
			return &free, errors.New("尚未设置出款费率")
		}

		if direction == model.ALL {
			if free.InFree == 0 && free.OutFree == 0 {
				return &free, errors.New("尚未设置费率")
			}
			if free.InFree == 0 {
				return &free, errors.New("尚未设置入款费率")
			}
			if free.OutFree == 0 {
				return &free, errors.New("尚未设置出款费率")
			}
		}
		//缓存起来
		helper.NoterMap.Store(freeKey, &free)
		return &free, nil
	}

	free := res.(*model.Free)

	if direction == model.PAY_IN && free.InFree == 0 {
		return free, errors.New("尚未设置入款费率")
	}
	if direction == model.PAY_OUT && free.OutFree == 0 {
		return free, errors.New("尚未设置出款费率")
	}

	if direction == model.ALL {
		if free.InFree == 0 && free.OutFree == 0 {
			return free, errors.New("尚未设置费率")
		}
		if free.InFree == 0 {
			return free, errors.New("尚未设置入款费率")
		}
		if free.OutFree == 0 {
			return free, errors.New("尚未设置出款费率")
		}
	}

	return res.(*model.Free), nil
}
