package genshin

import (
	"encoding/json"
	"errors"
	"fmt"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const id = "genshin_privete"
const name = "genshin私服"

type Genshin struct {
	gorm.Model
	Uid string
	QQ  int64
}

var genshinDb *gorm.DB

// initialize 初始化
func initialize(dbpath string) *gorm.DB {
	var err error
	if _, err = os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil
		}
		defer f.Close()
	}
	qdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		panic(err)
	}
	qdb.AutoMigrate(&Genshin{})
	return qdb
}

func init() {
	engine := control.Register(id, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "genshin私服\n" +
			"- 绑定私服UID xxxx\n" +
			"- 发送物品 ID 数量\n" +
			"- 物品查询 xxxx\n",
	})

	fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		path := engine.DataFolder() + "genshin.db"
		genshinDb = initialize(path)
		return true
	})

	engine.OnPrefix("绑定私服UID").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			suid := ctx.State["args"].(string)
			int64uid, err := strconv.ParseInt(suid, 10, 64)
			if suid == "" || int64uid < 100000000 || int64uid > 1000000000 || err != nil {
				//ctx.SendChain(message.Text("-请输入正确的uid"))
				return
			}
			exist := Genshin{}
			first := genshinDb.Model(Genshin{}).Where("qq = ?", ctx.Event.UserID).First(&exist)
			if first.Error != nil {
				return
			}
			if first.RowsAffected > 0 {
				// 更新绑定
				genshinDb.Model(Genshin{}).Where("qq = ?", ctx.Event.UserID).Update("uid", suid)
			} else {
				genshinDb.Create(&Genshin{QQ: ctx.Event.UserID, Uid: suid})
			}
		})
	engine.OnPrefix("发送物品").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			log.Info("收到发送物品消息")
			txt := ctx.State["args"].(string)
			log.Info("查询画：", txt)
			if txt != "" {
				split := strings.Split(txt, " ")
				if len(split) == 2 {
					ItemName := split[0]
					ItemNumber := split[1]
					sendGoods(ctx, ItemName, ItemNumber)
				}
			}
		})
}

type TokenRes struct {
	Token string `json:"token"`
}

func login(uid string, password string) string {
	url := "http://127.0.0.1:4001/login"
	method := "POST"

	m := map[string]interface{}{
		"username": uid, "password": password, "remember": true,
	}
	marshal, _ := json.Marshal(m)
	payload := strings.NewReader(string(marshal))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	tokenRes := TokenRes{}
	err = json.Unmarshal(body, &tokenRes)
	if err != nil {
		return ""
	}
	return tokenRes.Token
	//fmt.Println(string(body))
}

func getUidByQQ(qq int64) (string, error) {
	genshin := Genshin{}
	first := genshinDb.Model(Genshin{}).Where("qq = ?", qq).First(&genshin)
	if first.Error != nil || first.RowsAffected == 0 {
		return "", errors.New("未绑定UID")
	}
	return genshin.Uid, nil
}

func getGoodsIdByGoodsName(goodsName string) string {
	return goodsName
}

func sendGoods(ctx *zero.Ctx, goodsName string, goodsNumber string) {
	// 0. 获取GoodsName是否存在
	goodsId := getGoodsIdByGoodsName(goodsName)
	if len(goodsId) == 0 {
		ctx.SendChain(message.Text("物品名不存在: " + goodsName))
	}

	// 1. 获取当前用户的UID
	uid, err := getUidByQQ(ctx.Event.UserID)
	if err != nil {
		ctx.SendChain(message.Text(err.Error()))
	}
	// 3. 执行login获取token
	password := "user_password"
	if uid == "100000002" {
		password = "abc123_"
	}
	token := login(uid, password)
	if len(token) == 0 {
		ctx.SendChain(message.Text("登录失败"))
	}
	// 4. 执行发送物品
	parseInt, err := strconv.ParseInt(goodsNumber, 64, 10)
	if err != nil {
		ctx.SendChain(message.Text("物品数量不正确: " + goodsNumber))
	}
	sendGoodsByToken(ctx, uid, goodsId, parseInt, token)
}

func sendGoodsByToken(ctx *zero.Ctx, uid string, goodsId string, goodsNumber int64, token string) {
	url := "http://127.0.0.1:4001/api/item-name"
	method := "POST"

	requestBody := map[string]interface{}{
		"uid":    uid,
		"itemId": goodsId,
		"number": goodsNumber,
	}
	marshal, err2 := json.Marshal(requestBody)
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	payload := strings.NewReader(string(marshal))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(body))
		ctx.SendChain(message.Text("发送成功"))
	} else {
		ctx.SendChain(message.Text("发送失败"))
	}
}
