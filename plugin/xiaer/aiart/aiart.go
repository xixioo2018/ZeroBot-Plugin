package aiart

import (
	"encoding/json"
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/database/baidu"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
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
			"- 画一张**\n",
	})
	engine.OnRegex(`^画一张\s?(.*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			if strings.Contains(keyword, ",") {
				split := strings.Split(keyword, ",")
				if len(split) == 2 {
					style := split[0]
					text := split[1]
					art(ctx, style, text)
				}
			} else {
				art(ctx, "", keyword)
			}
		})
	//engine.OnPrefix("画一张").SetBlock(true).Limit(ctxext.LimitByGroup).
	//	Handle(func(ctx *zero.Ctx) {
	//		art(ctx)
	//	})
}

func art(ctx *zero.Ctx, style, text string) {
	token, err := getAccessToken()
	if err != nil {
		ctx.SendChain(message.Text(
			err.Error(),
		))
	}
	taskId, err := applyPic(token, style, text)
	if err != nil {
		ctx.SendChain(message.Text(
			err.Error(),
		))
	}

	go func() {
		for {
			result, isSuccess, err2 := pollingResult(token, taskId)
			if err2 != nil || !isSuccess {
				time.Sleep(5 * time.Second)
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
	var result PicRes
	if err = json.Unmarshal(rspBody, &result); err != nil {
		return 0, err
	}
	return result.Data.TaskId, nil
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

	if result.Data.Waiting == "0" {
		return &result.Data, true, nil
	} else {
		return nil, false, nil
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
