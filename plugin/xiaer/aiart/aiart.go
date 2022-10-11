package aiart

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/database/baidu"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const id = "aiart"
const name = "Ai绘画"

func init() {
	engine := control.Register(id, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "Ai绘画\n" +
			"- 画一张[style]的[xx] \n" +
			"- style: 古风 油画 水彩画 卡通画 二次元 浮世绘 蒸汽波艺术 low poly 像素风格 概念艺术 未来主义 赛博朋克 写实风格 洛丽塔风格 巴洛克风格 超现实主义\n",
	})
	engine.OnPrefix("画一张").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			log.Info("收到查询消息")
			txt := ctx.State["args"].(string)
			log.Info("查询画：", txt)
			if txt != "" {
				if strings.Contains(txt, "的") {
					split := strings.Split(txt, "的")
					if len(split) >= 2 {
						style := split[0]
						text := ""
						for index, word := range split {
							if index > 0 {
								text += word
							}
						}
						art(ctx, style, text)
					}
				} else {
					art(ctx, "二次元", txt)
				}
			}
		})
}

func art(ctx *zero.Ctx, style, text string) {
	log.Info("开始绘画: ", style, text)
	token, err := getAccessToken()
	if err != nil {
		ctx.SendChain(message.Text(
			err.Error(),
		))
		return
	}
	log.Info("AccessToken: ", token)
	taskId, err := applyPic(token, style, text)
	if err != nil {
		ctx.SendChain(message.Text(
			err.Error(),
		))
		return
	}
	ctx.SendChain(message.Text("已开始绘画，请稍后约30s，请勿重复发送!!!"))
	log.Info("taskId: ", taskId)

	go func() {
		count := 10
		for {
			if count <= 0 {
				break
			}
			result, isSuccess, err2 := pollingResult(token, taskId)
			if !isSuccess {
				if err2 != nil {
					ctx.SendChain(message.Text(
						err2.Error(),
					))
				}
				time.Sleep(10 * time.Second)
				count--
				continue
			}
			m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text(text))}
			for i, imgData := range result.ImgUrls {
				print(i)
				m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image(imgData.Image)))
			}

			if id := ctx.Send(m).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
			}
			break
		}
	}()
}

func getAccessToken() (string, error) {
	body := make(map[string]string)
	ak, sk := baidu.GetAkSk()
	body["grant_type"] = "client_credentials"
	body["client_id"] = ak
	body["client_secret"] = sk
	marshal, _ := json.Marshal(body)
	reqBody := strings.NewReader(string(marshal))
	tokenReq, _ := http.NewRequest("POST", "https://wenxin.baidu.com/moduleApi/portal/api/oauth/token", reqBody)
	tokenReq.Header.Add("Content-Type", "application/json")

	httpRsp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return "", err
	}
	defer httpRsp.Body.Close()

	rspBody, err := ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return "", err
	}
	var result TokenRes
	if err = json.Unmarshal(rspBody, &result); err != nil {
		return "", err
	}
	return result.Data, nil
}

func applyPic(accessToken string, style string, text string) (int, error) {
	body := make(map[string]string)
	body["style"] = style
	body["text"] = text
	marshal, _ := json.Marshal(body)
	reqBody := strings.NewReader(string(marshal))
	url := fmt.Sprintf("https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernievilg/v1/txt2img?access_token=%s", accessToken)
	tokenReq, _ := http.NewRequest("POST", url, reqBody)
	tokenReq.Header.Add("Content-Type", "application/json")

	httpRsp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return 0, err
	}
	defer httpRsp.Body.Close()

	rspBody, err := ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return 0, err
	}
	log.Info("申请结果: ", string(rspBody))
	var result PicRes
	if err = json.Unmarshal(rspBody, &result); err != nil {
		return 0, err
	}
	if result.Code == 0 {
		return result.Data.TaskId, nil
	} else {
		return 0, errors.New(result.Msg)
	}
}

func pollingResult(accessToken string, taskId int) (*QueryData, bool, error) {
	body := make(map[string]int)
	body["taskId"] = taskId
	marshal, _ := json.Marshal(body)
	reqBody := strings.NewReader(string(marshal))
	url := fmt.Sprintf("https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernievilg/v1/getImg?access_token=%s", accessToken)
	tokenReq, _ := http.NewRequest("POST", url, reqBody)
	tokenReq.Header.Add("Content-Type", "application/json")

	httpRsp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return nil, false, err
	}
	defer httpRsp.Body.Close()

	rspBody, err := ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return nil, false, err
	}
	var result QueryRes
	if err = json.Unmarshal(rspBody, &result); err != nil {
		return nil, false, err
	}
	log.Info("查询结果", string(rspBody))

	if result.Code == 0 {
		if result.Data.Waiting == "0" {
			return &result.Data, true, nil
		} else {
			return nil, false, nil
		}
	} else {
		return nil, false, errors.New(result.Msg)
	}
}

type TokenRes struct {
	Code int
	Msg  string
	Data string
}

type ApplyTask struct {
	TaskId    int
	RequestId string
}

type PicRes struct {
	Code int
	Msg  string
	Data ApplyTask
}

type QueryData struct {
	Img     string `json:"img"`
	Waiting string `json:"waiting"`
	ImgUrls []struct {
		Image string      `json:"image"`
		Score interface{} `json:"score"`
	} `json:"imgUrls"`
	CreateTime string `json:"createTime"`
	RequestId  string `json:"requestId"`
	Style      string `json:"style"`
	Text       string `json:"text"`
	Resolution string `json:"resolution"`
	TaskId     int    `json:"taskId"`
	Status     int    `json:"status"`
}

type QueryRes struct {
	Code int
	Msg  string
	Data QueryData
}
