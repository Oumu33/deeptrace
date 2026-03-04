package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cprobe/catpaw/config"
	"github.com/cprobe/catpaw/logger"
	"github.com/cprobe/catpaw/types"
)

type DingTalkNotifier struct {
	cfg    *config.DingTalkConfig
	client *http.Client
}

func NewDingTalkNotifier(cfg *config.DingTalkConfig) *DingTalkNotifier {
	return &DingTalkNotifier{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout),
		},
	}
}

func (d *DingTalkNotifier) Name() string { return "dingtalk" }

func (d *DingTalkNotifier) Forward(event *types.Event) bool {
	text := d.buildText(event)
	payload, err := json.Marshal(map[string]interface{}{
		"msgtype": "actionCard",
		"actionCard": map[string]interface{}{
			"title":          fmt.Sprintf("【DeepTrace】%s告警 - %s", event.EventStatus, event.Labels["plugin"]),
			"text":           text,
			"btnOrientation": "0",
		},
	})
	if err != nil {
		logger.Logger.Errorw("dingtalk: marshal fail", "event_key", event.AlertKey, "error", err.Error())
		return false
	}

	for attempt := 0; attempt <= d.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
			logger.Logger.Infow("dingtalk: retrying", "event_key", event.AlertKey, "attempt", attempt+1)
		}

		ok, retryable := d.doRequest(event.AlertKey, payload)
		if ok {
			return true
		}
		if !retryable {
			return false
		}
	}

	logger.Logger.Errorw("dingtalk: all retries exhausted", "event_key", event.AlertKey, "max_retries", d.cfg.MaxRetries)
	return false
}

func (d *DingTalkNotifier) doRequest(alertKey string, payload []byte) (ok bool, retryable bool) {
	webhookURL := d.cfg.Webhook
	if d.cfg.Secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := d.sign(timestamp, d.cfg.Secret)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", d.cfg.Webhook, timestamp, url.QueryEscape(sign))
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(payload))
	if err != nil {
		logger.Logger.Errorw("dingtalk: new request fail", "event_key", alertKey, "error", err.Error())
		return false, false
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := d.client.Do(req)
	if err != nil {
		logger.Logger.Errorw("dingtalk: do request fail", "event_key", alertKey, "error", err.Error())
		return false, true
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(io.LimitReader(res.Body, 4096))
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var resp struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		json.Unmarshal(body, &resp)
		if resp.ErrCode == 0 {
			logger.Logger.Infow("dingtalk: forward completed", "event_key", alertKey)
			return true, false
		}
		logger.Logger.Errorw("dingtalk: api error", "event_key", alertKey, "errcode", resp.ErrCode, "errmsg", resp.ErrMsg)
		return false, false
	}

	if res.StatusCode == 429 || res.StatusCode >= 500 {
		logger.Logger.Errorw("dingtalk: retryable error", "event_key", alertKey, "status", res.StatusCode)
		return false, true
	}

	logger.Logger.Errorw("dingtalk: non-retryable error", "event_key", alertKey, "status", res.StatusCode, "body", string(body))
	return false, false
}

func (d *DingTalkNotifier) sign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (d *DingTalkNotifier) buildText(event *types.Event) string {
	var sb strings.Builder

	plugin := event.Labels["plugin"]
	target := event.Labels["target"]
	host := event.Labels["from_hostname"]
	hostIP := event.Labels["from_hostip"]

	sb.WriteString(fmt.Sprintf("### 🚨 %s告警 - %s\n\n", event.EventStatus, plugin))
	sb.WriteString(fmt.Sprintf("告警级别：%s\n", statusEmoji(event.EventStatus)))

	currentValue := event.Attrs[types.AttrCurrentValue]
	if currentValue != "" {
		sb.WriteString(fmt.Sprintf("当前值：%s\n", currentValue))
	}

	sb.WriteString(fmt.Sprintf("主机：%s (%s)\n", host, hostIP))
	sb.WriteString(fmt.Sprintf("告警时间：%s\n", time.Unix(event.EventTime, 0).Format("2006-01-02 15:04:05")))

	if target != "" {
		sb.WriteString(fmt.Sprintf("目标：%s\n", target))
	}

	sb.WriteString("\n---\n\n")

	// Plugin info
	sb.WriteString(fmt.Sprintf("📊 插件：%s\n", plugin))

	// Description
	if event.Description != "" {
		sb.WriteString("\n---\n\n")
		desc := event.Description
		// Limit description length for DingTalk
		if len(desc) > 2000 {
			desc = desc[:2000] + "..."
		}
		sb.WriteString(fmt.Sprintf("📝 详情：\n%s\n", desc))
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString(fmt.Sprintf("alert_key: %s\n", event.AlertKey))
	sb.WriteString("\n由 DeepTrace 自动推送")

	return sb.String()
}
