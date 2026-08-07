package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeviceAuthSkillsEnforceWorkBuddyTurnBoundaries(t *testing.T) {
	root := filepath.Join("..", "..")
	auth := readDeviceAuthSkillFile(t, filepath.Join(root, "skills", "auth", "SKILL.md"))
	shared := readDeviceAuthSkillFile(t, filepath.Join(root, "skills", "contract-cli-shared", "SKILL.md"))

	for _, required := range []string{
		"`qr_code_data_uri`",
		"`show_widget`",
		"read_me(modules: \"diagram\")",
		"show_widget(title: \"智书合同授权二维码\"",
		"width:220px",
		"height:220px",
		"`expires_at_display`",
		"继续查询前，需要先完成智书合同授权。",
		"1. 点击蓝色链接「[打开授权页面](<verification_uri_complete>)」，或扫描本消息中的二维码",
		"2. 在 **<expires_at_display>** 前完成：手机号验证 + 企业确认授权",
		"3. 授权全部完成后，回复消息：**已授权**，我将立刻为你执行合同查询",
		"不得向用户展示 `expires_at` 的 RFC3339 原值",
		"不得使用反引号或代码样式展示时间",
		"present_files(files: [\"<qr_code_path>\"])",
		"data:image/png;base64",
		"不得把完整 Data URI 输出到最终回复正文",
		"只能调用一次授权展示工具",
		"禁止执行 `auth complete`、再次执行 `auth init`、业务命令、轮询或网络重试",
		"收到新的用户消息",
		"auth init --profile <profile> --output json --restart",
		"最终回复必须同时包含",
		"[打开授权页面](<verification_uri_complete>)",
		"WorkBuddy 主路径",
		"豆包继续将 `qr_code_path` 对应的 PNG 作为图片展示",
		"降级为可点击授权链接和 `qr_code_path` 对应的 PNG 产物卡片",
		"不得只返回授权链接或只返回二维码",
		"二维码内联展示失败，请点击图片卡片或授权链接",
		"不得声称二维码已经展示",
		"Refresh Token 返回 `invalid_grant` 时也必须先询问用户",
		"先执行 `auth status --profile <profile> --as user`",
		"根据真实状态选择复用现有会话、普通 `auth init` 或带 `--restart` 的 `auth init`",
		"auth status` 不支持 `--output",
		"豆包工作任务运行在云端任务环境",
		"macOS Keychain",
		"必须提供 `SKILL_SESSION_WORKSPACE`",
		"WorkBuddy 运行在客户本机",
	} {
		if !strings.Contains(auth, required) {
			t.Fatalf("auth skill missing required WorkBuddy rule %q", required)
		}
	}
	if strings.Contains(auth, "当前 user 身份未授权") {
		t.Fatal("auth skill still contains the old technical authorization copy")
	}
	if strings.Contains(auth, "固定为 320×320") {
		t.Fatal("auth skill still renders the WorkBuddy QR code at the oversized 320x320 display size")
	}
	for _, forbidden := range []string{"DOUBAO_SESSION_ID", "DOUBAO_TASK_ID", "豆包本地 Skill 必须提供", "豆包 Skill 只暴露固定子命令和结构化参数"} {
		if strings.Contains(auth, forbidden) {
			t.Fatalf("auth skill still documents removed Doubao local runtime contract %q", forbidden)
		}
	}
	for _, ambiguous := range []string{"扫描上方二维码", "扫描下方二维码"} {
		if strings.Contains(auth, ambiguous) {
			t.Fatalf("auth skill still uses layout-dependent QR copy %q", ambiguous)
		}
	}
	for _, required := range []string{
		"WorkBuddy 使用 `show_widget` 内联展示二维码",
		"豆包继续使用 `qr_code_path`",
		"Skill / 模型层不得重试任何 OAuth 命令",
		"CLI 内部仅对 `auth init` 的 TCP `dial` 失败自动重试一次",
		"请求已发送后的超时、HTTP 5xx、响应中断或解析失败不重试",
		"`auth complete`、Token 刷新和撤销始终不自动重试",
	} {
		if !strings.Contains(shared, required) {
			t.Fatalf("shared skill missing required OAuth retry boundary %q", required)
		}
	}
	for _, required := range []string{
		"CLI 内部的单次安全重试不算第二次 `auth init` 命令",
		"禁止额外执行 `curl`、`auth status` 或其他探测命令",
		"明确询问用户是否重新发起授权",
	} {
		if !strings.Contains(auth, required) {
			t.Fatalf("auth skill missing required init failure boundary %q", required)
		}
	}
}

func readDeviceAuthSkillFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
