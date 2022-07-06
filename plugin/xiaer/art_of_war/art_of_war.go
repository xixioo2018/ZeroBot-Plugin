package art_of_war

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/database/mongo"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	logger "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const id = "ART_OF_WAR"
const name = "兵法"

func init() {
	engine := control.Register(id, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "兵法\n" +
			"- 查兵法[xxx]",
	})
	engine.OnPrefix("兵法").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.State["args"].(string)
			if txt != "" {
				ctx.SendChain(message.Text("发送以下内容 \n\n 查兵法XX"))
			}
		})
	engine.OnPrefix("查兵法").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.State["args"].(string)
			if txt != "" {
				result := searchArtOfWar(txt)
				ctx.SendChain(message.Text(result))
			}
		})
	engine.OnPrefix("加兵法:").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.State["args"].(string)
			if txt != "" {
				if ctx.Event.UserID == 383541400 {
					result := addArtOfWar(txt)
					ctx.SendChain(message.Text(result))
				} else {
					ctx.SendChain(message.Text("请联系机器人管理员：383541400添加兵法"))
				}
				ctx.SendChain(message.Text("https://buhuibaidu.me/?s=" + url.QueryEscape(txt)))
			}
		})
}

type LogMessage struct {
	Content string
}

func addArtOfWar(content string) string {
	search := content[10:]
	fmt.Println(search)
	if len(search) <= 2 {
		return "请输入要添加的数据"
	}
	split := strings.Split(search, "\n")
	if len(split) >= 1 {
		for _, data := range split {
			datas := strings.Split(data, "##")
			if len(datas) == 2 {
				war := ArtOfWar{Question: datas[0], Answer: datas[1]}
				updateFromMongo(war)
			}
		}
		return "添加成功"
	} else {
		return "格式错误"
	}
}

func searchArtOfWar(content string) string {
	search := content[9:]
	fmt.Println(search)
	if len(search) < 2 {
		return "请输入题目"
	} else {
		fromMongo := searchFromMongo(search)
		if len(fromMongo) > 0 {
			result := "查询结果: \n\n"
			for index, artOfWar := range fromMongo {
				if index > 3 {
					continue
				}
				result += fmt.Sprintf(" 题目: %s \n 答案: %s \n\n", artOfWar.Question, artOfWar.Answer)
			}
			return result
		} else {
			// 从大嘴巴处查询结果
			fromDZB, err := searchFromDZB(search)
			if err != nil {
				logger.Error("调用错误", err)
				return "暂无结果"
			} else {
				result := "查询结果: 来源DZB\n\n"
				for index, artOfWar := range fromDZB {
					if index > 3 {
						continue
					}
					result += fmt.Sprintf(" 题目: %s \n 答案: %s \n", artOfWar.Question, artOfWar.Answer)
					updateFromMongo(artOfWar)
				}
				return result
			}
		}
	}
}

func searchFromDZB(search string) ([]ArtOfWar, error) {
	body := DzbQuestionBody{TableName: "B_STAR_ANSWER", Page: 1, Limit: 10, Filters: []DzbQuestionFilter{
		{FieldName: "SUBJECT,TRUE_ANSWER,FALSE_ANSWER", Type: "string", Compared: "like", FilterValue: search},
		{FieldName: "STATUS", Type: "date", Compared: "=", FilterValue: "已启用"},
	}, OrderByField: "CRT_TIME", IsDesc: 1}
	marshal, _ := json.Marshal(body)
	reqBody := strings.NewReader(string(marshal))
	dzbQuestionReq, _ := http.NewRequest("POSt", "https://www.dazuiba.top:8005/api/_search/postSearch", reqBody)
	dzbQuestionReq.Header.Add("Content-Type", "application/json")
	dzbQuestionReq.Header.Add("Accept-Encoding", "gzip, deflate, br")
	dzbQuestionReq.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36 MicroMessenger/7.0.9.501 NetType/WIFI MiniProgramEnv/Windows WindowsWechat`)
	dzbQuestionReq.Header.Add("Content-Type", "application/json")

	httpRsp, err := http.DefaultClient.Do(dzbQuestionReq)
	if err != nil {
		return nil, err
	}
	defer httpRsp.Body.Close()

	rspBody, err := ioutil.ReadAll(httpRsp.Body)

	if err != nil {
		return nil, err
	}

	var result DzbResponseData
	if err = json.Unmarshal(rspBody, &result); err != nil {
		return nil, err
	}

	if result.ISSUCCESS {
		wars := make([]ArtOfWar, 0)
		for _, table := range result.Table {
			wars = append(wars, ArtOfWar{Question: table.SUBJECT, Answer: table.CHHANSWER})
		}
		return wars, nil
	}
	return nil, errors.New("调用请求失败")
}

type DzbResponseData struct {
	Id        string                 `json:"$id"`
	Table     []DzbResponseDataTable `json:"Table"`
	ISSUCCESS bool                   `json:"IS_SUCCESS"`
	MSG       string                 `json:"MSG"`
}

type DzbResponseDataTable struct {
	KID         int    `json:"KID"`
	STATUS      string `json:"STATUS"`
	SUBJECT     string `json:"SUBJECT"`
	TRUEANSWER  string `json:"TRUE_ANSWER"`
	FALSEANSWER string `json:"FALSE_ANSWER"`
	FILLER      string `json:"FILLER"`
	FILLERID    int    `json:"FILLER_ID"`
	ISDELETE    bool   `json:"IS_DELETE"`
	CRTTIME     string `json:"CRT_TIME"`
	CODE        string `json:"CODE"`
	TYPE        string `json:"TYPE"`
	JINLIANSWER string `json:"JINLI_ANSWER"`
	CHHANSWER   string `json:"CHH_ANSWER"`
}

type DzbQuestionBody struct {
	TableName    string              `json:"tableName"`
	Page         int                 `json:"page"`
	Limit        int                 `json:"limit"`
	Filters      []DzbQuestionFilter `json:"filters"`
	OrderByField string              `json:"orderByField"`
	IsDesc       int                 `json:"isDesc"`
}

type DzbQuestionFilter struct {
	FieldName   string `json:"fieldName"`
	Type        string `json:"type"`
	Compared    string `json:"compared"`
	FilterValue string `json:"filterValue"`
}

type ArtOfWar struct {
	Question string
	Answer   string
}

func searchFromMongo(question string) []ArtOfWar {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := mongo.Collection("query.art.war")
	defer cancel()
	cur, err := coll.Find(ctx, bson.M{
		"question": bson.M{"$regex": primitive.Regex{Pattern: ".*" + question + ".*", Options: "i"}}, // : groupCode
	}, options.Find())

	defer cur.Close(ctx)
	artOfWars := make([]ArtOfWar, 0)
	for cur.Next(context.TODO()) {
		var artOfWar ArtOfWar
		err = cur.Decode(&artOfWar)
		if err != nil {
			panic(err)
		}
		artOfWars = append(artOfWars, artOfWar)
	}
	return artOfWars
}

func updateFromMongo(artOfWar ArtOfWar) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := mongo.Collection("query.art.war")
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"question": artOfWar.Question},
		bson.M{"$set": bson.M{
			"question": artOfWar.Question,
			"answer":   artOfWar.Answer,
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		logger.Error("新增数据失败")
		panic(err)
	}
}
