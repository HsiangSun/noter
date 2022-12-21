package telgram

import (
	tb "gopkg.in/telebot.v3"
	"noter/model"
	"strings"
)

var (
	BINDING    = "绑定"
	START_NOTE = "记账"
	END_NOTE   = "清账"
)

func OnTextMessage(c tb.Context) error {
	msg := c.Text()

	if msg == BINDING {
		return Binding(c)
	}

	if msg == START_NOTE {
		return StartNote(c)
	}

	if strings.HasPrefix(msg, "入款") {
		return PayIn(c)
	}

	if strings.HasPrefix(msg, "出款") {
		return PayOut(c)
	}

	if strings.HasPrefix(msg, "设置入款汇率") {
		return SetInRate(c)
	}

	if strings.HasPrefix(msg, "设置出款汇率") {
		return SetOutRate(c)
	}

	if strings.Contains(msg, "+") {
		return Record(c, model.PAY_IN)
	}

	if strings.Contains(msg, "-") {
		return Record(c, model.PAY_OUT)
	}

	return nil
}
