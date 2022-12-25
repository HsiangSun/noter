package telgram

import (
	tb "gopkg.in/telebot.v3"
	"noter/model"
	"strings"
)

var (
	START_NOTE = "记账"
	END_NOTE   = "清账"
	HELP       = "帮助"
)

func OnTextMessage(c tb.Context) error {
	msg := c.Text()

	if msg == START_NOTE {
		return StartNote(c)
	}

	if msg == HELP {
		return Help(c)
	}

	if msg == END_NOTE {
		return clean(c)
	}

	if strings.HasPrefix(msg, "设置入款汇率") {
		return SetInRate(c)
	}

	if strings.HasPrefix(msg, "设置币种") {
		return SetCurrency(c)
	}

	if strings.HasPrefix(msg, "授权") {
		return SetAdmin(c)
	}

	if strings.HasPrefix(msg, "设置出款汇率") {
		return SetOutRate(c)
	}

	if strings.HasPrefix(msg, "设置入款费率") {
		return SetInFree(c)
	}

	if strings.HasPrefix(msg, "设置出款费率") {
		return SetOutFree(c)
	}

	if strings.Contains(msg, "+") {
		return Record(c, model.PAY_IN)
	}

	if strings.Contains(msg, "-") {
		return Record(c, model.PAY_OUT)
	}

	return nil
}
