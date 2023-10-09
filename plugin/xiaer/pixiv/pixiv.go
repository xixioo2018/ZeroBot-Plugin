package pixiv

import (
	"encoding/json"
	"fmt"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const id = "pixiv"
const name = "pixiv搜图"

func init() {
	engine := control.Register(id, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "pixiv搜图\n" +
			"- 抽老公[style] \n" +
			"- pixiv搜图[style] \n" +
			"- style: 帅哥 腹肌\n",
	})
	engine.OnPrefix("抽老公").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			log.Info("收到查询消息")
			txt := ctx.State["args"].(string)
			log.Info("查询画：", txt)
			if txt != "" {
				txt = "帅哥 " + txt
				search(ctx, txt)
			}
		})
	engine.OnPrefix("pixiv搜图").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			log.Info("收到查询消息")
			txt := ctx.State["args"].(string)
			log.Info("查询画：", txt)
			if txt != "" {
				search(ctx, txt)
			}
		})
}

func getProxyClient() (*http.Client, error) {
	proxyURL, err := url.Parse("http://192.168.2.83:10811")
	// 创建一个SOCKS5代理拨号器
	if err != nil {
		return nil, err
	}

	// 创建一个HTTP客户端，并配置代理
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	return client, nil
}

func search(ctx *zero.Ctx, text string) {
	log.Info("开始查询 ", text)
	reqUrl := "https://www.pixiv.net/ajax/search/illustrations/" + text + "?word=" + text + "&order=date_d&mode=all&p=1&s_mode=s_tag&type=illust_and_ugoira&lang=zh"
	method := "GET"

	client, err := getProxyClient()
	if err != nil {
		ctx.SendChain(message.Text(
			err.Error(),
		))
		return
	}
	req, err := http.NewRequest(method, reqUrl, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	pixivResponse := PixivResponse{}
	err = json.Unmarshal(body, &pixivResponse)
	if err != nil {
		ctx.SendChain(message.Text(
			err.Error(),
		))
	}
	length := len(pixivResponse.Body.Illust.Data)
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(length)
	data := pixivResponse.Body.Illust.Data[randomNum]
	getPicDataFromProxy(ctx, data.Url)
}

func getPicDataFromProxy(ctx *zero.Ctx, picUrl string) {
	client, err := getProxyClient()
	if err != nil {
		ctx.SendChain(message.Text(
			err.Error(),
		))
		return
	}
	req, err := http.NewRequest("GET", picUrl, nil)
	req.Header.Add("Referer", "https://www.pixiv.net/")

	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx.SendChain(message.ImageBytes(body))
}

type PixivResponse struct {
	Error bool `json:"error"`
	Body  struct {
		Illust struct {
			Data []struct {
				Id                      string      `json:"id"`
				Title                   string      `json:"title"`
				IllustType              int         `json:"illustType"`
				XRestrict               int         `json:"xRestrict"`
				Restrict                int         `json:"restrict"`
				Sl                      int         `json:"sl"`
				Url                     string      `json:"url"`
				Description             string      `json:"description"`
				Tags                    []string    `json:"tags"`
				UserId                  string      `json:"userId"`
				UserName                string      `json:"userName"`
				Width                   int         `json:"width"`
				Height                  int         `json:"height"`
				PageCount               int         `json:"pageCount"`
				IsBookmarkable          bool        `json:"isBookmarkable"`
				BookmarkData            interface{} `json:"bookmarkData"`
				Alt                     string      `json:"alt"`
				TitleCaptionTranslation struct {
					WorkTitle   interface{} `json:"workTitle"`
					WorkCaption interface{} `json:"workCaption"`
				} `json:"titleCaptionTranslation"`
				CreateDate      time.Time `json:"createDate"`
				UpdateDate      time.Time `json:"updateDate"`
				IsUnlisted      bool      `json:"isUnlisted"`
				IsMasked        bool      `json:"isMasked"`
				AiType          int       `json:"aiType"`
				ProfileImageUrl string    `json:"profileImageUrl"`
			} `json:"data"`
			Total          int `json:"total"`
			LastPage       int `json:"lastPage"`
			BookmarkRanges []struct {
				Min *int        `json:"min"`
				Max interface{} `json:"max"`
			} `json:"bookmarkRanges"`
		} `json:"illust"`
		Popular struct {
			Recent    []interface{} `json:"recent"`
			Permanent []interface{} `json:"permanent"`
		} `json:"popular"`
		RelatedTags    []string                     `json:"relatedTags"`
		TagTranslation map[string]map[string]string `json:"tagTranslation"`
		ZoneConfig     struct {
			Header struct {
				Url string `json:"url"`
			} `json:"header"`
			Footer struct {
				Url string `json:"url"`
			} `json:"footer"`
			Infeed struct {
				Url string `json:"url"`
			} `json:"infeed"`
		} `json:"zoneConfig"`
		ExtraData struct {
			Meta struct {
				Title              string `json:"title"`
				Description        string `json:"description"`
				Canonical          string `json:"canonical"`
				AlternateLanguages struct {
					Ja string `json:"ja"`
					En string `json:"en"`
				} `json:"alternateLanguages"`
				DescriptionHeader string `json:"descriptionHeader"`
			} `json:"meta"`
		} `json:"extraData"`
	} `json:"body"`
}
