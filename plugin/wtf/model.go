package wtf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

/* JS path getter for https://wtf.hiigara.net/ranking
a = document.getElementById("testList").getElementsByTagName("a")
s = ""
for(i=0; i<a.length; i++) {
    s += "\"" + a[i].innerText + "\":\"" + a[i].href + "\",\n";
}
*/

const apiprefix = "https://wtf.hiigara.net/api/run/"

type wtf struct {
	name string
	path string
}

var table = [...]*wtf{
	{"你的意义是什么?", "mRIFuS"},
	{"【ABO】性別和信息素", "KXyy9"},
	{"测测cp", "ZoGXQd"},
	{"xxx和xxx的關係是？", "L4HfA"},
	{"在JOJO世界，你的替身会是什么？", "lj0a8o"},
	{"稱號產生器", "titlegen"},
	{"成分报告", "2PCeo1"},
	{"測驗你跟你的朋友是攻/受", "LkQXO3"},
	{"测试两人的关系？", "uwjQQt"},
	{"【Fate系列】當你成為了從者 2.0", "LHStH2"},
	{"想不到自己未來要做什麼工作嗎?", "D1agGa"},
	{"(σﾟ∀ﾟ)σ名字產生器", "LNxXq7"},
	{"人設生產器", "LBtPu5"},
	{"測驗你在ABO世界的訊息素", "SwmdU"},
	{"爱是什么", "llpBEY"},
	{"測測你和哪位名人相似？", "RHQeXu"},
	{"S/M测试", "Ga47oZ"},
	{"测测你是谁", "aV1AEi"},
	{"取個綽號吧", "LTkyUy"},
	{"什麼都不是", "vyrSCb"},
	{"今天中午吃什麼", "LdS4K6"},
	{"測試你的中二稱號", "LwUmQ6"},
	{"神奇海螺", "Lon1h7"},
	{"ABO測試", "H1Tgd"},
	{"女主角姓名產生器", "MsQBTd"},
	{"您是什么人", "49PwSd"},
	{"如果你成为了干员", "ok5e7n"},
	{"abo人设生成~", "Di8enA"},
	{"✡你的命運✡塔羅占卜🔮", "ohCzID"},
	{"小說大綱生產器", "Lnstjz"},
	{"他会喜欢你吗？", "pezX3a"},
	{"抽签！你明年的今天会干什么", "IF31kS"},
	{"如果你是受，會是哪種受呢？", "Dr6zpF"},
	{"cp文梗", "vEO2KD"},
	{"您是什么人？", "TQ5qyl"},
	{"你成為......的機率", "g0uoBL"},
	{"ABO性別與信息素", "KFPju"},
	{"異國名稱產生器(國家、人名、星球...)", "OBpu4"},
	{"對方到底喜不喜歡你", "JSLoZC"},
	{"【脑叶公司】测一测你在脑叶公司的经历", "uPBhjC"},
	{"当你成为魔法少女", "7ZiGcJ"},
	{"你是yyds吗?", "SpBnCa"},
	{"○○喜歡你嗎？", "S6Uceo"},
	{"测测你的sm属性", "dOtcO5"},
	{"你/妳究竟是攻還是受呢?", "RXALH"},
	{"神秘藏书阁", "tDRyET"},
	{"中午吃什么？", "L0Wsis"},
	{"十年后，你cp的结局是", "VUwnXQ"},
	{"高维宇宙与常数的你", "6Zql97"},
	{"色色的東東", "o2eg74"},
	{"文章標題產生器", "Ky25WO"},
	{"你的成績怎麼樣", "6kZv69"},
	{"智能SM偵測器ヾ(*ΦωΦ)ツ", "9pY6HQ"},
	{"你的使用注意事項", "La4Gir"},
	{"戀愛指數", "Jsgz0"},
	{"测试你今晚拉的屎", "N8dbcL"},
	{"成為情侶的機率ᶫᵒᵛᵉᵧₒᵤ♥", "eDURch"},
	{"他對你...", "CJxHMf"},
	{"你的明日方舟人际关系", "u5z4Mw"},
	{"日本姓氏產生器", "JJ5Ctb"},
	{"當你轉生到了異世界，你將成為...", "FTpwK"},
	{"魔幻世界大穿越2.0", "wUATOq"},
	{"未來男朋友", "F3dSV"},
	{"ABO與信息素", "KFOGA"},
	{"你必將就這樣一事無成啊アホ", "RWw9oX"},
	{"用習慣舉手的方式測試你的戀愛運!<3", "wv5bzA"},
	{"攻受", "RaKmY"},
	{"你和你喜歡的人的微h寵溺段子XD", "LdQqGz"},
	{"我的藝名", "LBaTx"},
	{"你是什麼神？", "LqZORE"},
	{"你的起源是什麼？", "HXWwC"},
	{"測你喜歡什麼", "Sue5g2"},
	{"看看朋友的秘密", "PgKb8r"},
	{"你在動漫裡的名字", "Lz82V7"},
	{"小說男角名字產生器", "LyGDRr"},
	{"測試短文", "S48yA"},
	{"我們兩人在一起的機率......", "LBZbgE"},
	{"創造小故事", "Kjy3AS"},
	{"你的另外一個名字", "LuyYQA"},
	{"與你最匹配的攻君屬性 ！？", "I7pxy"},
	{"英文全名生產器(女)", "HcYbq"},
	{"BL文章生產器", "LBZMO"},
	{"輕小說書名產生器", "NFucA"},
	{"長相評分", "2cQSDP"},
	{"日本名字產生器（女孩子）", "JRiKv"},
	{"中二技能名產生器", "Ky1BA"},
	{"抽籤", "XqxfuH"},
	{"你的蘿莉控程度全國排名", "IIWh9k"},
}

func newWtf(index int) *wtf {
	if index >= 0 && index < len(table) {
		return table[index]
	}
	return nil
}

type result struct {
	Text string `json:"text"`
	// Path string `json:"path"`
	Ok  bool   `json:"ok"`
	Msg string `json:"msg"`
}

func (w *wtf) predict(names ...string) (string, error) {
	name := ""
	for _, n := range names {
		name += "/" + url.QueryEscape(n)
	}
	u := apiprefix + w.path + name
	r, err := GetDataFromProxy(u)
	if err != nil {
		return "", err
	}
	re := new(result)
	err = json.Unmarshal(r, re)
	if err != nil {
		return "", err
	}
	if re.Ok {
		return "> " + w.name + "\n" + re.Text, nil
	}
	return "", errors.New(re.Msg)
}

// GetDataFromProxy 获取数据
func GetDataFromProxy(urlStr string) (data []byte, err error) {
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

	var response *http.Response
	response, err = client.Get(urlStr)
	if err == nil {
		if response.StatusCode != http.StatusOK {
			s := fmt.Sprintf("status code: %d", response.StatusCode)
			err = errors.New(s)
			return
		}
		data, err = io.ReadAll(response.Body)
		response.Body.Close()
	}
	return
}
