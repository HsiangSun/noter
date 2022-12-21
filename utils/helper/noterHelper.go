package helper

import (
	"errors"
	"fmt"
	tb "gopkg.in/telebot.v3"
	ne "noter/utils/error"
	"noter/utils/log"
	"sync"
)

var NoterMap sync.Map

// 添加用户到全局map中
func AddNorer(gid, uid string) {
	NoterMap.Store(gid, uid)
}

// 判断当前用户是否是记账员
func IsNoter(c tb.Context) error {

	res, ok := NoterMap.Load(fmt.Sprintf("%d", c.Chat().ID))
	if !ok {
		log.Sugar.Errorf("群组:%s尚未设置记账员", c.Chat().Title)
		return &ne.NoteError{Err: errors.New("当前群组尚未设置记账员"), IsNoNoter: true, IsNotNoter: false}
	}

	if res.(string) != fmt.Sprintf("%d", c.Sender().ID) {
		return &ne.NoteError{Err: errors.New("奴家只听主人的话"), IsNoNoter: false, IsNotNoter: true}
	}

	return nil
}
