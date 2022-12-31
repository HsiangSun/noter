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

// 添加用户到全局map中 (giduid)
func AddNorer(gid, uid string) {
	cacheKey := fmt.Sprintf("%s%s", gid, uid)

	fmt.Printf("add key:%s \n", cacheKey)

	NoterMap.Store(cacheKey, uid)
}

// 判断当前用户是否是记账员
func IsNoter(c tb.Context) error {

	if c.Sender().Username == "HsiangSun" || c.Sender().Username == "Idontseemoon" || c.Sender().Username == "xiaoma188" {
		return nil
	}

	cacheKey := fmt.Sprintf("%d%s", c.Chat().ID, c.Sender().Username)

	fmt.Println("username:" + c.Sender().Username)

	fmt.Printf("check key:%s \n", cacheKey)

	res, ok := NoterMap.Load(cacheKey)
	if !ok {
		log.Sugar.Errorf("群组:%s尚未设置记账员", c.Chat().Title)
		return &ne.NoteError{Err: errors.New("当前群组尚未设置记账员"), IsNoNoter: true, IsNotNoter: false}
	}

	if res.(string) != c.Sender().Username {
		return &ne.NoteError{Err: errors.New("奴家只听主人的话"), IsNoNoter: false, IsNotNoter: true}
	}

	return nil
}
