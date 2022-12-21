package telgram

import (
	tb "gopkg.in/telebot.v3"
	"noter/utils/helper"
	"noter/utils/log"
)

func CheckNoter(c tb.Context) error {
	err := helper.IsNoter(c)
	if err != nil {
		//if ins, ok := err.(*ne.NoteError); ok {
		//
		//}
		tErr := c.Delete()
		if tErr != nil {
			log.Sugar.Errorf("Send mesage error:%s", tErr.Error())
		}
	}
	return err
}
