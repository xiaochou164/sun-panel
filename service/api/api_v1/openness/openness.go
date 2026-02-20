package openness

import (
	"encoding/json"
	"io"
	"net/http"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/global"
	"sun-panel/lib/cmn/systemSetting"
	"time"

	"github.com/gin-gonic/gin"
)

type Openness struct {
}

func (a *Openness) LoginConfig(c *gin.Context) {
	cfg := systemSetting.ApplicationSetting{}
	if err := global.SystemSetting.GetValueByInterface(systemSetting.SYSTEM_APPLICATION, &cfg); err != nil {
		apiReturn.Error(c, "配置查询失败："+err.Error())
		return
	}
	apiReturn.SuccessData(c, gin.H{
		"loginCaptcha": cfg.LoginCaptcha,
		"register":     cfg.Register,
	})
}

func (a *Openness) GetDisclaimer(c *gin.Context) {
	if content, err := global.SystemSetting.GetValueString(systemSetting.DISCLAIMER); err != nil {
		global.SystemSetting.Set(systemSetting.DISCLAIMER, "")
		apiReturn.SuccessData(c, "")
		return
	} else {
		apiReturn.SuccessData(c, content)
	}
}

func (a *Openness) GetAboutDescription(c *gin.Context) {
	if content, err := global.SystemSetting.GetValueString(systemSetting.WEB_ABOUT_DESCRIPTION); err != nil {
		global.SystemSetting.Set(systemSetting.WEB_ABOUT_DESCRIPTION, "")
		apiReturn.SuccessData(c, "")
		return
	} else {
		apiReturn.SuccessData(c, content)
	}
}

func (a *Openness) GetBingWallpaper(c *gin.Context) {
	n := c.DefaultQuery("n", "1")
	if n != "1" && n != "2" && n != "3" && n != "4" && n != "5" && n != "6" && n != "7" && n != "8" {
		n = "1"
	}
	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=" + n + "&mkt=zh-CN"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		apiReturn.Error(c, "请求失败:"+err.Error())
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		apiReturn.Error(c, "获取 Bing 壁纸失败:"+err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	}
	if err := json.Unmarshal(body, &result); err != nil || len(result.Images) == 0 {
		apiReturn.Error(c, "解析 Bing 壁纸失败")
		return
	}

	urls := make([]string, len(result.Images))
	for i, img := range result.Images {
		urls[i] = "https://www.bing.com" + img.URL
	}
	apiReturn.SuccessData(c, gin.H{"urls": urls, "url": urls[0]})
}
