package txt2img

import (
	"encoding/json"
	"errors"
	"fmt"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io/ioutil"
	"net/http"
	"strings"
)

const id = "txt2img"
const name = "三次元文转图"

func init() {
	engine := control.Register(id, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "文转图\n" +
			"- 文转图 \n[提示] \n[负面提示]" +
			"- 提示: <lora:koreanDollLikeness_v10:0.8>, best quality, ultra high res, (photorealistic:1.4), 1woman, sleeveless white button shirt, black skirt, black choker, cute, (Kpop idol), (aegyo sal:1), (silver hair:1), ((puffy eyes)), looking at viewer, peace sign\n" +
			"- 负面提示: paintings, sketches, (worst quality:2), (low quality:2), (normal quality:2), lowres, normal quality, ((monochrome)), ((grayscale)), skin spots, acnes, skin blemishes, age spot, glans, nsfw, nipples" +
			"- 可选：自带参数<lora:koreanDollLikeness_v10:0.8>, (photorealistic:1.4)",
	})
	engine.OnPrefix("文转图").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			log.Info("收到查询消息")
			txt := ctx.State["args"].(string)
			log.Info("查询画：", txt)
			if txt != "" {
				split := strings.Split(txt, "\n")
				var prompt string
				var negativePrompt = "paintings, sketches, (worst quality:2), (low quality:2), (normal quality:2), lowres, normal quality, ((monochrome)), ((grayscale)), skin spots, acnes, skin blemishes, age spot, glans, nsfw, nipples"
				if len(split) == 2 {
					prompt = split[1]
				} else if len(split) == 3 {
					prompt = split[1]
					negativePrompt = split[2]
				}
				if !strings.Contains(prompt, "DollLikeness_v") {
					prompt = "<lora:koreanDollLikeness_v10:0.8>, " + prompt
				}
				if !strings.Contains(prompt, "photorealistic") {
					prompt += ", (photorealistic:1.4)"
				}

				art(ctx, prompt, negativePrompt)
			}
		})
}

func art(ctx *zero.Ctx, prompt, negativePrompt string) {
	log.Info("开始绘画: ", prompt, negativePrompt)
	base64Img, err := applyPic(prompt, negativePrompt)
	if err != nil {
		ctx.SendChain(message.Text(
			err.Error(),
		))
		return
	}
	ctx.SendChain(message.Text("已开始绘画，请稍后约30s，请勿重复发送!!!"))

	m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text(prompt))}
	m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image("base64://"+base64Img)))

	if id := ctx.Send(m).ID(); id == 0 {
		ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
	}
}

func applyPic(prompt, negativePrompt string) (string, error) {
	body := map[string]interface{}{
		"enable_hr":                            false,
		"denoising_strength":                   0,
		"firstphase_width":                     0,
		"firstphase_height":                    0,
		"hr_scale":                             2,
		"hr_upscaler":                          "string",
		"hr_second_pass_steps":                 0,
		"hr_resize_x":                          0,
		"hr_resize_y":                          0,
		"prompt":                               prompt,
		"seed":                                 -1,
		"subseed":                              -1,
		"subseed_strength":                     0,
		"seed_resize_from_h":                   -1,
		"seed_resize_from_w":                   -1,
		"sampler_name":                         "Euler a",
		"batch_size":                           1,
		"n_iter":                               1,
		"steps":                                20,
		"cfg_scale":                            7,
		"width":                                512,
		"height":                               768,
		"restore_faces":                        false,
		"tiling":                               false,
		"negative_prompt":                      negativePrompt,
		"eta":                                  0,
		"s_churn":                              0,
		"s_tmax":                               0,
		"s_tmin":                               0,
		"s_noise":                              1,
		"override_settings_restore_afterwards": true,
		"sampler_index":                        "Euler",
	}
	marshal, _ := json.Marshal(body)
	reqBody := strings.NewReader(string(marshal))
	url := fmt.Sprintf("http://ddns.xiaer.ml:7860/sdapi/v1/txt2img")
	tokenReq, _ := http.NewRequest("POST", url, reqBody)
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
	var result PicResult
	if err = json.Unmarshal(rspBody, &result); err != nil {
		return "", err
	}
	if len(result.Images) > 0 {
		return result.Images[0], nil
	} else {
		return "", errors.New("生成失败")
	}
}

type PicResult struct {
	Images     []string `json:"images"`
	Parameters struct {
		EnableHr          bool        `json:"enable_hr"`
		DenoisingStrength float64     `json:"denoising_strength"`
		FirstphaseWidth   int         `json:"firstphase_width"`
		FirstphaseHeight  int         `json:"firstphase_height"`
		HrScale           float64     `json:"hr_scale"`
		HrUpscaler        string      `json:"hr_upscaler"`
		HrSecondPassSteps int         `json:"hr_second_pass_steps"`
		HrResizeX         int         `json:"hr_resize_x"`
		HrResizeY         int         `json:"hr_resize_y"`
		Prompt            string      `json:"prompt"`
		Styles            interface{} `json:"styles"`
		Seed              int         `json:"seed"`
		Subseed           int         `json:"subseed"`
		SubseedStrength   float64     `json:"subseed_strength"`
		SeedResizeFromH   int         `json:"seed_resize_from_h"`
		SeedResizeFromW   int         `json:"seed_resize_from_w"`
		SamplerName       string      `json:"sampler_name"`
		BatchSize         int         `json:"batch_size"`
		NIter             int         `json:"n_iter"`
		Steps             int         `json:"steps"`
		CfgScale          float64     `json:"cfg_scale"`
		Width             int         `json:"width"`
		Height            int         `json:"height"`
		RestoreFaces      bool        `json:"restore_faces"`
		Tiling            bool        `json:"tiling"`
		NegativePrompt    string      `json:"negative_prompt"`
		Eta               float64     `json:"eta"`
		SChurn            float64     `json:"s_churn"`
		STmax             float64     `json:"s_tmax"`
		STmin             float64     `json:"s_tmin"`
		SNoise            float64     `json:"s_noise"`
		OverrideSettings  struct {
		} `json:"override_settings"`
		OverrideSettingsRestoreAfterwards bool          `json:"override_settings_restore_afterwards"`
		ScriptArgs                        []interface{} `json:"script_args"`
		SamplerIndex                      string        `json:"sampler_index"`
		ScriptName                        interface{}   `json:"script_name"`
	} `json:"parameters"`
	Info string `json:"info"`
}
