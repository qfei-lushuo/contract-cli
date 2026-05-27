package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type helpTopic struct {
	Name     string
	Summary  string
	Usage    []string
	Commands []helpCommand
	Flags    []helpFlag
	Examples []string
	Notes    []string
}

type helpCommand struct {
	Name        string
	Description string
}

type helpFlag struct {
	Name        string
	Description string
}

func isHelpRequest(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if args[0] == "help" {
		return true
	}
	for _, arg := range args {
		if isHelpFlag(arg) {
			return true
		}
	}
	return false
}

func isHelpFlag(arg string) bool {
	switch arg {
	case "--help", "-h", "-help":
		return true
	default:
		return false
	}
}

func resolveHelpTopic(args []string) (helpTopic, error) {
	registry := helpRegistry()
	path := normalizeHelpPath(args)
	if len(path) == 0 {
		return registry["contract-cli"], nil
	}

	key := strings.Join(path, " ")
	if topic, ok := registry[key]; ok {
		return topic, nil
	}

	bestKey := ""
	bestLen := 0
	for i := len(path); i > 0; i-- {
		candidate := strings.Join(path[:i], " ")
		if _, ok := registry[candidate]; ok {
			bestKey = candidate
			bestLen = i
			break
		}
	}

	if bestKey != "" {
		topic := registry[bestKey]
		if len(topic.Commands) == 0 {
			return topic, nil
		}
		if bestLen < len(path) {
			unknown := strings.Join(path[:bestLen+1], " ")
			return helpTopic{}, unknownHelpTopicError(unknown)
		}
	}

	return helpTopic{}, unknownHelpTopicError(key)
}

func normalizeHelpPath(args []string) []string {
	start := 0
	if len(args) > 0 && args[0] == "help" {
		start = 1
	}

	path := make([]string, 0, len(args)-start)
	for _, arg := range args[start:] {
		if isHelpFlag(arg) {
			continue
		}
		path = append(path, arg)
	}
	return path
}

func unknownHelpTopicError(topic string) error {
	return fmt.Errorf("unknown help topic %q; run `contract-cli help` to list available commands", topic)
}

func renderHelp(writer io.Writer, topic helpTopic) error {
	if _, err := fmt.Fprintf(writer, "Name:\n  %s\n", topic.Name); err != nil {
		return err
	}
	if topic.Summary != "" {
		if _, err := fmt.Fprintf(writer, "\n%s\n", topic.Summary); err != nil {
			return err
		}
	}
	if len(topic.Usage) > 0 {
		if _, err := fmt.Fprintln(writer, "\nUsage:"); err != nil {
			return err
		}
		for _, usage := range topic.Usage {
			if _, err := fmt.Fprintf(writer, "  %s\n", usage); err != nil {
				return err
			}
		}
	}
	if len(topic.Commands) > 0 {
		if _, err := fmt.Fprintln(writer, "\nCommands:"); err != nil {
			return err
		}
		if err := renderHelpCommands(writer, topic.Commands); err != nil {
			return err
		}
	}
	if len(topic.Flags) > 0 {
		if _, err := fmt.Fprintln(writer, "\nFlags:"); err != nil {
			return err
		}
		if err := renderHelpFlags(writer, topic.Flags); err != nil {
			return err
		}
	}
	if len(topic.Examples) > 0 {
		if _, err := fmt.Fprintln(writer, "\nExamples:"); err != nil {
			return err
		}
		for _, example := range topic.Examples {
			if _, err := fmt.Fprintf(writer, "  %s\n", example); err != nil {
				return err
			}
		}
	}
	if len(topic.Notes) > 0 {
		if _, err := fmt.Fprintln(writer, "\nNotes:"); err != nil {
			return err
		}
		for _, note := range topic.Notes {
			if _, err := fmt.Fprintf(writer, "  - %s\n", note); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(writer)
	return err
}

func renderHelpCommands(writer io.Writer, commands []helpCommand) error {
	table := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	for _, command := range commands {
		if _, err := fmt.Fprintf(table, "  %s\t%s\n", command.Name, command.Description); err != nil {
			return err
		}
	}
	return table.Flush()
}

func renderHelpFlags(writer io.Writer, flags []helpFlag) error {
	table := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	for _, flag := range flags {
		if _, err := fmt.Fprintf(table, "  %s\t%s\n", flag.Name, flag.Description); err != nil {
			return err
		}
	}
	return table.Flush()
}

func helpRegistry() map[string]helpTopic {
	topCommands := []helpCommand{
		{"contract-cli config add [flags]", "初始化或更新 profile"},
		{"contract-cli auth login [flags]", "登录 user 或 bot 身份"},
		{"contract-cli auth status [flags]", "查看授权状态"},
		{"contract-cli auth logout [flags]", "登出指定身份"},
		{"contract-cli auth use [flags]", "切换默认业务身份"},
		{"contract-cli version", "查看版本信息"},
		{"contract-cli skills list", "列出内置 Agent skills"},
		{"contract-cli skills install [flags]", "安装内置 Agent skills"},
		{"contract-cli update check [flags]", "检查 npm 远端版本"},
		{"contract-cli contract <subcommand> [flags]", "合同结构化命令"},
		{"contract-cli payment <subcommand> [flags]", "付款结构化命令"},
		{"contract-cli mdm vendor <subcommand> [flags]", "交易方主数据命令"},
		{"contract-cli mdm legal <subcommand> [flags]", "法人主体主数据命令"},
		{"contract-cli mdm fields list [flags]", "字段配置查询"},
	}

	registry := map[string]helpTopic{
		"contract-cli": {
			Name:    "contract-cli",
			Summary: "合同开放平台命令行工具。",
			Usage: []string{
				"contract-cli <command> [flags]",
				"contract-cli help <command path>",
				"contract-cli <command path> --help",
			},
			Commands: topCommands,
			Notes: []string{
				"使用 `contract-cli <command> --help` 查看具体命令参数。",
				"JSON 请求体统一使用 --input-file 或 --data；真实文件上传使用 --file。",
			},
		},
		"version": {
			Name:    "version",
			Summary: "查看当前 CLI 版本、commit 和构建时间。",
			Usage:   []string{"contract-cli version", "contract-cli --version"},
			Examples: []string{
				"contract-cli version",
				"contract-cli --version",
			},
		},
	}

	addConfigHelp(registry)
	addAuthHelp(registry)
	addSkillsHelp(registry)
	addUpdateHelp(registry)
	addContractHelp(registry)
	addPaymentHelp(registry)
	addMDMHelp(registry)
	return registry
}

func addConfigHelp(registry map[string]helpTopic) {
	registry["config"] = helpTopic{
		Name:  "config",
		Usage: []string{"contract-cli config <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli config add [flags]", "初始化或更新 profile"},
		},
	}
	registry["config add"] = helpTopic{
		Name:    "config add",
		Summary: "初始化或更新 profile，并写入开放平台、user OAuth 和 bot token endpoint 配置。",
		Usage:   []string{"contract-cli config add [flags]"},
		Flags: []helpFlag{
			{"--env <prod>", "环境预设；当前仅支持 prod，默认 prod"},
			{"--name <profile>", "profile 名称，默认 contract"},
			{"--resource-metadata-url <url>", "覆盖 protected resource metadata 地址"},
			{"--redirect-url <url>", "覆盖 OAuth callback 地址"},
			{"--scope <scopes>", "覆盖默认 OAuth scopes，多个 scope 用空格分隔"},
		},
		Examples: []string{
			"contract-cli config add --env prod --name contract",
		},
	}
}

func addAuthHelp(registry map[string]helpTopic) {
	registry["auth"] = helpTopic{
		Name:  "auth",
		Usage: []string{"contract-cli auth <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli auth login [flags]", "登录 user 或 bot 身份"},
			{"contract-cli auth status [flags]", "查看授权状态"},
			{"contract-cli auth logout [flags]", "登出指定身份"},
			{"contract-cli auth use [flags]", "切换默认业务身份"},
		},
	}
	registry["auth login"] = helpTopic{
		Name:    "auth login",
		Summary: "登录 user 或 bot 身份。user 走 OAuth 授权，bot 使用 appId/appSecret 换取 tenant_access_token。",
		Usage:   []string{"contract-cli auth login [flags]"},
		Flags: []helpFlag{
			{"--profile <name>", "profile 名称；不传使用当前 profile"},
			{"--as <user|bot>", "登录身份，默认 user"},
			{"--timeout <duration>", "user OAuth 等待时间，默认 3m"},
			{"--no-open-browser", "只打印授权 URL，不自动打开浏览器"},
			{"--app-id <id>", "bot app id；bot 登录时可通过 flag/env/secrets 提供"},
			{"--app-secret <secret>", "bot app secret；不会输出到日志"},
		},
		Examples: []string{
			"contract-cli auth login --profile contract --as user",
			"contract-cli auth login --profile contract --as bot --app-id <id> --app-secret <secret>",
		},
		Notes: []string{
			"bot 凭证优先级：flag > env > 已保存 secrets。",
			"bot 登录成功后保存 token，并把 default_identity 切到 bot。",
		},
	}
	registry["auth status"] = helpTopic{
		Name:    "auth status",
		Summary: "查看某个 profile 的 user 或 bot 身份状态。",
		Usage:   []string{"contract-cli auth status [flags]"},
		Flags: []helpFlag{
			{"--profile <name>", "profile 名称；不传使用当前 profile"},
			{"--as <user|bot>", "查看身份，默认 user"},
		},
		Examples: []string{
			"contract-cli auth status --profile contract --as user",
			"contract-cli auth status --profile contract --as bot",
		},
	}
	registry["auth logout"] = helpTopic{
		Name:    "auth logout",
		Summary: "清理指定身份的 token。",
		Usage:   []string{"contract-cli auth logout [flags]"},
		Flags: []helpFlag{
			{"--profile <name>", "profile 名称；不传使用当前 profile"},
			{"--as <user|bot>", "登出身份，默认 user"},
		},
		Examples: []string{
			"contract-cli auth logout --profile contract --as user",
			"contract-cli auth logout --profile contract --as bot",
		},
		Notes: []string{
			"bot logout 只清空 bot token，保留 appId/appSecret。",
		},
	}
	registry["auth use"] = helpTopic{
		Name:    "auth use",
		Summary: "切换 profile 默认业务身份。",
		Usage:   []string{"contract-cli auth use [flags]"},
		Flags: []helpFlag{
			{"--profile <name>", "profile 名称；不传使用当前 profile"},
			{"--as <user|bot>", "默认业务身份，默认 user"},
		},
		Examples: []string{
			"contract-cli auth use --profile contract --as bot",
		},
	}
}

func addSkillsHelp(registry map[string]helpTopic) {
	registry["skills"] = helpTopic{
		Name:  "skills",
		Usage: []string{"contract-cli skills <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli skills list", "列出内置 Agent skills"},
			{"contract-cli skills install [flags]", "安装内置 Agent skills"},
		},
		Notes: []string{
			"推荐跨平台安装方式：npx skills add qfeius/contract-cli -y -g。",
		},
	}
	registry["skills list"] = helpTopic{
		Name:    "skills list",
		Summary: "列出当前二进制内置的 Agent skills。",
		Usage:   []string{"contract-cli skills list"},
		Examples: []string{
			"contract-cli skills list",
		},
	}
	registry["skills install"] = helpTopic{
		Name:    "skills install",
		Summary: "把当前二进制内置 skills 安装到本机 Codex skills 目录。",
		Usage:   []string{"contract-cli skills install [flags]"},
		Flags: []helpFlag{
			{"--target <dir>", "安装目标目录；默认 $CODEX_HOME/skills 或 ~/.codex/skills"},
			{"--force", "覆盖已存在的同名 skill；默认跳过"},
		},
		Examples: []string{
			"contract-cli skills install",
			"contract-cli skills install --target ~/.codex/skills",
			"contract-cli skills install --force",
		},
	}
}

func addUpdateHelp(registry map[string]helpTopic) {
	registry["update"] = helpTopic{
		Name:  "update",
		Usage: []string{"contract-cli update <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli update check [flags]", "检查 npm 远端版本"},
		},
	}
	registry["update check"] = helpTopic{
		Name:    "update check",
		Summary: "检查 npm 远端是否存在可升级版本。",
		Usage:   []string{"contract-cli update check [flags]"},
		Flags: []helpFlag{
			{"--channel <latest|beta>", "npm dist-tag；正式版通常使用 latest，不传时根据当前版本推断"},
		},
		Examples: []string{
			"contract-cli update check",
			"contract-cli update check --channel latest",
		},
		Notes: []string{
			"交互终端下普通命令最多每 30 分钟自动检查一次。",
			"可设置 CONTRACT_CLI_NO_UPDATE_CHECK=1 关闭自动检查。",
		},
	}
}

func addAPIHelp(registry map[string]helpTopic) {
	registry["api"] = helpTopic{
		Name:  "api",
		Usage: []string{"contract-cli api <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli api call <METHOD> <PATH> [flags]", "原始开放平台接口调用"},
		},
	}
	registry["api call"] = helpTopic{
		Name:    "api call",
		Summary: "对开放平台任意相对路径发起原始调用。",
		Usage:   []string{"contract-cli api call <METHOD> <PATH> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags(), []helpFlag{
			{"--header \"Key: Value\"", "追加 HTTP header，可重复传入"},
		}),
		Examples: []string{
			"contract-cli api call GET /open-apis/contract/v1/mcp/config/config_list --profile contract --as user",
			"contract-cli api call POST /open-apis/mdm/v1/vendors --profile contract --as bot --data '{\"foo\":\"bar\"}'",
		},
		Notes: []string{
			"PATH 必须是相对路径，且以 /open-apis/ 开头。",
			"/open-apis/contract/v1/mcp/... 会被视为 user-only，显式 --as bot 会报错。",
		},
	}
}

func addContractHelp(registry map[string]helpTopic) {
	registry["contract"] = helpTopic{
		Name:  "contract",
		Usage: []string{"contract-cli contract <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli contract search [flags]", "搜索合同"},
			{"contract-cli contract get <contract-id> [flags]", "获取合同详情"},
			{"contract-cli contract sync-user-groups [flags]", "同步用户分组"},
			{"contract-cli contract text <contract-id> [flags]", "获取合同文本"},
			{"contract-cli contract create [flags]", "创建合同"},
			{"contract-cli contract upload-file [flags]", "上传合同文件"},
			{"contract-cli contract submit <contract-id> [flags]", "bot 身份提交合同"},
			{"contract-cli contract resubmit <contract-id> [flags]", "bot 身份重新提交合同"},
			{"contract-cli contract patch <contract-id> [flags]", "bot 身份更新合同"},
			{"contract-cli contract download-file <file-id> [flags]", "bot 身份下载合同相关文件"},
			{"contract-cli contract delete <contract-id> [flags]", "bot 身份删除草稿合同"},
			{"contract-cli contract print-file [flags]", "bot 身份生成合同打印文件"},
			{"contract-cli contract share <subcommand> [flags]", "bot 身份查询合同分享记录"},
			{"contract-cli contract cooperation <resource> <subcommand> [flags]", "bot 身份查询合同协商信息"},
			{"contract-cli contract approval <subcommand> [flags]", "bot 身份操作审批流程"},
			{"contract-cli contract category list [flags]", "列出合同分类"},
			{"contract-cli contract template <subcommand> [flags]", "模板相关命令"},
			{"contract-cli contract enum list [flags]", "查询枚举值"},
		},
	}
	registry["contract search"] = helpTopic{
		Name:    "contract search",
		Summary: "搜索合同，按当前身份自动路由 user MCP 或 bot 开放平台接口。",
		Usage:   []string{"contract-cli contract search [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags(), pageFlags(), []helpFlag{
			{"--contract-number <number>", "按合同编号搜索，会合并进 JSON 请求体"},
		}),
		Examples: []string{
			"contract-cli contract search --profile contract --as user --input-file search.json",
			"contract-cli contract search --profile contract --as bot --data '{\"contract_number\":\"CN-001\"}'",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/contracts/search",
			"bot: /open-apis/contract/v1/contracts/search",
			"--input-file / --data 可选；查询 flag 会合并进 JSON body。",
		},
	}
	registry["contract get"] = helpTopic{
		Name:    "contract get",
		Summary: "获取合同详情，按当前身份自动路由。",
		Usage:   []string{"contract-cli contract get <contract-id> [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli contract get <contract-id> --profile contract --as user",
			"contract-cli contract get <contract-id> --profile contract --as bot --user-id-type employee_id",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/contracts/{contract_id}",
			"bot: /open-apis/contract/v1/contracts/{contract_id}",
		},
	}
	registry["contract sync-user-groups"] = helpTopic{
		Name:    "contract sync-user-groups",
		Summary: "同步合同用户分组。",
		Usage:   []string{"contract-cli contract sync-user-groups [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli contract sync-user-groups --profile contract --as user",
			"contract-cli contract sync-user-groups --profile contract --as bot --user-id ou_xxx",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/contracts/user-groups/sync",
			"bot: /open-apis/contract/v1/contracts/user-groups/sync",
		},
	}
	registry["contract text"] = helpTopic{
		Name:    "contract text",
		Summary: "获取合同文本。",
		Usage:   []string{"contract-cli contract text <contract-id> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--full-text", "获取完整文本"},
			{"--offset <n>", "文本偏移量"},
			{"--limit <n>", "文本长度限制"},
		}),
		Examples: []string{
			"contract-cli contract text <contract-id> --profile contract --as user --full-text",
			"contract-cli contract text <contract-id> --profile contract --as bot --offset 0 --limit 1000",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/contracts/{contract_id}/text",
			"bot: /open-apis/contract/v1/contracts/{contract_id}/text",
		},
	}
	registry["contract create"] = helpTopic{
		Name:    "contract create",
		Summary: "创建合同，请求体必须是 JSON object。",
		Usage:   []string{"contract-cli contract create --input-file <path>|--data <json> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli contract create --profile contract --input-file create.json",
			"contract-cli contract create --profile contract --as bot --data '{\"title\":\"demo\",\"create_user_id\":\"ou_xxx\"}'",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/contracts",
			"bot: /open-apis/contract/v1/contracts",
			"bot 创建合同时，请调用方自行在 JSON body 中提供 create_user_id。",
		},
	}
	registry["contract upload-file"] = helpTopic{
		Name:    "contract upload-file",
		Summary: "上传合同相关文件，返回后端原始 JSON，重点关注 data.file_id。",
		Usage:   []string{"contract-cli contract upload-file --file <path> --file-type <type> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--file <path>", "必填，本地待上传文件路径"},
			{"--file-type <type>", "必填，文件类型，例如 text、attachment、scan"},
			{"--file-name <name>", "可选，上传给后端的文件名；默认使用本地文件名"},
		}),
		Examples: []string{
			"contract-cli contract upload-file --profile contract --as user --file ./合同正文.docx --file-type text",
			"contract-cli contract upload-file --profile contract --as bot --file ./附件.pdf --file-type attachment --file-name 附件.pdf",
		},
		Notes: []string{
			"user/bot 均走 POST /open-apis/contract/v1/files/upload。",
			"请求使用 multipart/form-data，字段为 file_name、file_type、file。",
			"本地文件必须存在、是普通文件，大小 <= 200MB。",
			"不接受 --input-file / --data；这两个参数只用于 JSON 请求体。",
		},
	}
	registry["contract submit"] = helpTopic{
		Name:    "contract submit",
		Summary: "bot 身份提交合同。",
		Usage:   []string{"contract-cli contract submit <contract-id> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli contract submit <contract-id> --profile contract --as bot",
			"contract-cli contract submit <contract-id> --profile contract --as bot --data '{\"comment\":\"ok\"}'",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 POST /open-apis/contract/v1/contracts/{contract_id}/submit。",
			"--input-file / --data 可选；不传时不发送请求体。",
		},
	}
	registry["contract resubmit"] = helpTopic{
		Name:    "contract resubmit",
		Summary: "bot 身份重新提交合同。",
		Usage:   []string{"contract-cli contract resubmit <contract-id> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli contract resubmit <contract-id> --profile contract --as bot",
			"contract-cli contract resubmit <contract-id> --profile contract --as bot --input-file resubmit.json",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 POST /open-apis/contract/v1/contracts/{contract_id}/resubmit。",
			"--input-file / --data 可选；不传时不发送请求体。",
		},
	}
	registry["contract patch"] = helpTopic{
		Name:    "contract patch",
		Summary: "bot 身份更新合同，请求体必须是 JSON。",
		Usage:   []string{"contract-cli contract patch <contract-id> --input-file <path>|--data <json> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli contract patch <contract-id> --profile contract --as bot --input-file patch.json",
			"contract-cli contract patch <contract-id> --profile contract --as bot --data '{\"title\":\"demo\"}'",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 PATCH /open-apis/contract/v1/contracts/{contract_id}。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["contract download-file"] = helpTopic{
		Name:    "contract download-file",
		Summary: "bot 身份下载合同相关文件。",
		Usage:   []string{"contract-cli contract download-file <file-id> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--output-file <path>", "保存到指定文件；不传时默认拉起保存文件弹窗"},
			{"--force", "覆盖已存在的 --output-file"},
		}),
		Examples: []string{
			"contract-cli contract download-file <file-id> --profile contract --as bot",
			"contract-cli contract download-file <file-id> --profile contract --as bot --output-file ./contract.pdf",
			"contract-cli contract download-file <file-id> --profile contract --as bot --raw > contract.pdf",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/files/{file_id}。",
			"默认拉起保存文件弹窗；无 GUI/远程/CI 环境推荐显式传 --output-file。",
			"--raw 会把文件内容写到 stdout，不打印额外提示。",
			"不实现 dowload-file 拼写别名。",
		},
	}
	registry["contract delete"] = helpTopic{
		Name:    "contract delete",
		Summary: "bot 身份删除草稿合同。",
		Usage:   []string{"contract-cli contract delete <contract-id> [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli contract delete <contract-id> --profile contract --as bot",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 DELETE /open-apis/contract/v1/contracts/{contract_id}。",
			"命令直接删除，不额外要求 --yes。",
		},
	}
	registry["contract print-file"] = helpTopic{
		Name:    "contract print-file",
		Summary: "bot 身份生成合同打印文件，请求体必须是 JSON。",
		Usage:   []string{"contract-cli contract print-file --input-file <path>|--data <json> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli contract print-file --profile contract --as bot --input-file print-file.json",
			"contract-cli contract print-file --profile contract --as bot --data '{\"contract_id\":\"<contract-id>\"}'",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 POST /open-apis/contract/v1/files。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["contract share"] = helpTopic{
		Name:  "contract share",
		Usage: []string{"contract-cli contract share <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli contract share get <contract-id> [flags]", "bot 身份查询合同分享记录"},
		},
	}
	registry["contract share get"] = helpTopic{
		Name:    "contract share get",
		Summary: "bot 身份查询合同分享记录。",
		Usage:   []string{"contract-cli contract share get <contract-id> [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli contract share get <contract-id> --profile contract --as bot",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/contracts/{contract_id}/share_records。",
		},
	}
	registry["contract cooperation"] = helpTopic{
		Name:  "contract cooperation",
		Usage: []string{"contract-cli contract cooperation <resource> <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli contract cooperation link get <contract-id> [flags]", "bot 身份查询合同协商邀请链接"},
			{"contract-cli contract cooperation record get <contract-id> [flags]", "bot 身份查询合同协商操作记录"},
		},
	}
	registry["contract cooperation link get"] = helpTopic{
		Name:    "contract cooperation link get",
		Summary: "bot 身份查询合同协商邀请链接。",
		Usage:   []string{"contract-cli contract cooperation link get <contract-id> [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli contract cooperation link get <contract-id> --profile contract --as bot",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/contracts/{contract_id}/cooperation_link。",
		},
	}
	registry["contract cooperation record get"] = helpTopic{
		Name:    "contract cooperation record get",
		Summary: "bot 身份查询合同协商操作记录信息。",
		Usage:   []string{"contract-cli contract cooperation record get <contract-id> [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli contract cooperation record get <contract-id> --profile contract --as bot",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/contracts/{contract_id}/cooperation_record_info。",
		},
	}
	registry["contract approval"] = helpTopic{
		Name:  "contract approval",
		Usage: []string{"contract-cli contract approval <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli contract approval start <process-instance-id> [flags]", "bot 身份发起流程审批"},
			{"contract-cli contract approval get <process-instance-id> [flags]", "bot 身份查询审批实例详情"},
		},
	}
	registry["contract approval start"] = helpTopic{
		Name:    "contract approval start",
		Summary: "bot 身份发起流程审批，请求体必须是 JSON。",
		Usage:   []string{"contract-cli contract approval start <process-instance-id> --input-file <path>|--data <json> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli contract approval start <process-instance-id> --profile contract --as bot --input-file approval.json",
			"contract-cli contract approval start <process-instance-id> --profile contract --as bot --data '{\"task_instance_id\":\"task-1\",\"command_type\":\"general\"}'",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 POST /open-apis/contract/v1/process_instances/{process_instance_id}/task_approval。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["contract approval get"] = helpTopic{
		Name:    "contract approval get",
		Summary: "bot 身份查询审批实例详情。",
		Usage:   []string{"contract-cli contract approval get <process-instance-id> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--notice-filter <filter>", "可选，审批实例详情查询参数 notice_filter"},
			{"--task-instance-filter <filter>", "可选，审批实例详情查询参数 task_instance_filter"},
		}),
		Examples: []string{
			"contract-cli contract approval get <process-instance-id> --profile contract --as bot",
			"contract-cli contract approval get <process-instance-id> --profile contract --as bot --notice-filter notice_filter --task-instance-filter task_instance_filter",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/process_instances/{process_instance_id}。",
			"不接受 --input-file / --data。",
		},
	}
	addContractNestedHelp(registry)
}

func addPaymentHelp(registry map[string]helpTopic) {
	registry["payment"] = helpTopic{
		Name:  "payment",
		Usage: []string{"contract-cli payment <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli payment create [flags]", "bot 身份创建付款申请"},
			{"contract-cli payment update <payment-id> [flags]", "bot 身份更新付款信息"},
			{"contract-cli payment get <payment-id> [flags]", "bot 身份查看付款信息"},
			{"contract-cli payment list [flags]", "bot 身份查询付款申请列表"},
			{"contract-cli payment plan <subcommand> [flags]", "bot 身份操作付款计划"},
			{"contract-cli payment record <subcommand> [flags]", "bot 身份操作付款记录"},
		},
	}
	registry["payment create"] = helpTopic{
		Name:    "payment create",
		Summary: "bot 身份创建付款申请，请求体必须是 JSON。",
		Usage:   []string{"contract-cli payment create --contract <contract-id> --input-file <path>|--data <json> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags(), []helpFlag{
			{"--contract <contract-id>", "必填，父级合同 ID"},
		}),
		Examples: []string{
			"contract-cli payment create --contract <contract-id> --profile contract --as bot --input-file payment.json",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 POST /open-apis/contract/v1/contracts/{contract_id}/payments。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["payment update"] = helpTopic{
		Name:    "payment update",
		Summary: "bot 身份更新付款信息，请求体必须是 JSON。",
		Usage:   []string{"contract-cli payment update <payment-id> --contract <contract-id> --input-file <path>|--data <json> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags(), []helpFlag{
			{"--contract <contract-id>", "必填，父级合同 ID"},
		}),
		Examples: []string{
			"contract-cli payment update <payment-id> --contract <contract-id> --profile contract --as bot --input-file payment-update.json",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 PATCH /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["payment get"] = helpTopic{
		Name:    "payment get",
		Summary: "bot 身份查看付款信息。",
		Usage:   []string{"contract-cli payment get <payment-id> --contract <contract-id> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--contract <contract-id>", "必填，父级合同 ID"},
		}),
		Examples: []string{
			"contract-cli payment get <payment-id> --contract <contract-id> --profile contract --as bot",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}。",
			"不接受 --input-file / --data。",
		},
	}
	registry["payment list"] = helpTopic{
		Name:    "payment list",
		Summary: "bot 身份查询付款申请列表。",
		Usage:   []string{"contract-cli payment list --contract <contract-id> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), pageFlags(), []helpFlag{
			{"--contract <contract-id>", "必填，父级合同 ID"},
		}),
		Examples: []string{
			"contract-cli payment list --contract <contract-id> --profile contract --as bot",
			"contract-cli payment list --contract <contract-id> --profile contract --as bot --page-size 10 --page-token next",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/contracts/{contract_id}/payments。",
			"不接受 --input-file / --data。",
		},
	}
	registry["payment plan"] = helpTopic{
		Name:  "payment plan",
		Usage: []string{"contract-cli payment plan <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli payment plan notify [flags]", "bot 身份同步付款记录"},
			{"contract-cli payment plan search [flags]", "bot 身份搜索付款计划"},
		},
	}
	registry["payment plan notify"] = helpTopic{
		Name:    "payment plan notify",
		Summary: "bot 身份同步付款记录，请求体必须是 JSON。",
		Usage:   []string{"contract-cli payment plan notify --input-file <path>|--data <json> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli payment plan notify --profile contract --as bot --input-file notify.json",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 POST /open-apis/contract/v1/payment/notify。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["payment plan search"] = helpTopic{
		Name:    "payment plan search",
		Summary: "bot 身份搜索付款计划，请求体必须是 JSON。",
		Usage:   []string{"contract-cli payment plan search --input-file <path>|--data <json> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli payment plan search --profile contract --as bot --input-file payment-plan-search.json",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 POST /open-apis/contract/v1/payments/search。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["payment record"] = helpTopic{
		Name:  "payment record",
		Usage: []string{"contract-cli payment record <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli payment record create [flags]", "bot 身份创建付款记录"},
			{"contract-cli payment record update <payment-record-id> [flags]", "bot 身份更新付款记录"},
			{"contract-cli payment record get <payment-record-id> [flags]", "bot 身份查询付款记录详情"},
			{"contract-cli payment record list [flags]", "bot 身份按付款计划查询付款记录"},
		},
	}
	registry["payment record create"] = helpTopic{
		Name:    "payment record create",
		Summary: "bot 身份创建付款记录，请求体必须是 JSON。",
		Usage:   []string{"contract-cli payment record create --contract <contract-id> --payment <payment-id> --input-file <path>|--data <json> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags(), []helpFlag{
			{"--contract <contract-id>", "必填，父级合同 ID"},
			{"--payment <payment-id>", "必填，父级付款 ID"},
		}),
		Examples: []string{
			"contract-cli payment record create --contract <contract-id> --payment <payment-id> --profile contract --as bot --input-file payment-record.json",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 POST /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}/payment_records。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["payment record update"] = helpTopic{
		Name:    "payment record update",
		Summary: "bot 身份更新付款记录，请求体必须是 JSON。",
		Usage:   []string{"contract-cli payment record update <payment-record-id> --contract <contract-id> --payment <payment-id> --input-file <path>|--data <json> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags(), []helpFlag{
			{"--contract <contract-id>", "必填，父级合同 ID"},
			{"--payment <payment-id>", "必填，父级付款 ID"},
		}),
		Examples: []string{
			"contract-cli payment record update <payment-record-id> --contract <contract-id> --payment <payment-id> --profile contract --as bot --input-file payment-record-update.json",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 PATCH /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}/payment_records/{payment_record_id}。",
			"--input-file / --data 必填且互斥。",
		},
	}
	registry["payment record get"] = helpTopic{
		Name:    "payment record get",
		Summary: "bot 身份查询付款记录详情。",
		Usage:   []string{"contract-cli payment record get <payment-record-id> --contract <contract-id> --payment <payment-id> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--contract <contract-id>", "必填，父级合同 ID"},
			{"--payment <payment-id>", "必填，父级付款 ID"},
		}),
		Examples: []string{
			"contract-cli payment record get <payment-record-id> --contract <contract-id> --payment <payment-id> --profile contract --as bot",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}/payment_records/{payment_record_id}。",
			"不接受 --input-file / --data。",
		},
	}
	registry["payment record list"] = helpTopic{
		Name:    "payment record list",
		Summary: "bot 身份根据付款计划 ID 查询付款记录。",
		Usage:   []string{"contract-cli payment record list --plan <payment-plan-uuid> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--plan <payment-plan-uuid>", "必填，付款计划 UUID"},
		}),
		Examples: []string{
			"contract-cli payment record list --plan <payment-plan-uuid> --profile contract --as bot",
		},
		Notes: []string{
			"bot-only: 当前仅支持 --as bot。",
			"走 GET /open-apis/contract/v1/contracts/payments/{payment_plan_uuid}/payment_records。",
			"不接受 --input-file / --data。",
		},
	}
}

func addContractNestedHelp(registry map[string]helpTopic) {
	registry["contract category"] = helpTopic{
		Name:  "contract category",
		Usage: []string{"contract-cli contract category <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli contract category list [flags]", "列出合同分类"},
		},
	}
	registry["contract category list"] = helpTopic{
		Name:    "contract category list",
		Summary: "列出合同分类。",
		Usage:   []string{"contract-cli contract category list [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--lang <lang>", "语言，例如 zh-CN"},
		}),
		Examples: []string{
			"contract-cli contract category list --profile contract",
			"contract-cli contract category list --profile contract --as bot --lang zh-CN",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/contract_categorys",
			"bot: /open-apis/contract/v1/contract_categorys",
		},
	}
	registry["contract template"] = helpTopic{
		Name:  "contract template",
		Usage: []string{"contract-cli contract template <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli contract template list [flags]", "列出模板"},
			{"contract-cli contract template get <template-id> [flags]", "获取模板详情"},
			{"contract-cli contract template instantiate [flags]", "创建模板实例"},
		},
	}
	registry["contract template list"] = helpTopic{
		Name:    "contract template list",
		Summary: "列出模板。",
		Usage:   []string{"contract-cli contract template list [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), pageFlags(), []helpFlag{
			{"--category-number <number>", "合同分类编号"},
		}),
		Examples: []string{
			"contract-cli contract template list --profile contract",
			"contract-cli contract template list --profile contract --as bot --category-number CAT-1 --page-size 20",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/templates",
			"bot: /open-apis/contract/v1/templates",
		},
	}
	registry["contract template get"] = helpTopic{
		Name:    "contract template get",
		Summary: "获取模板详情。",
		Usage:   []string{"contract-cli contract template get <template-id> [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli contract template get <template-id> --profile contract",
			"contract-cli contract template get <template-id> --profile contract --as bot --user-id ou_xxx",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/templates/{template_id}",
			"bot: /open-apis/contract/v1/templates/{template_id}",
		},
	}
	registry["contract template instantiate"] = helpTopic{
		Name:    "contract template instantiate",
		Summary: "创建模板实例，请求体必须是 JSON object。",
		Usage:   []string{"contract-cli contract template instantiate --input-file <path>|--data <json> [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), jsonBodyFlags()),
		Examples: []string{
			"contract-cli contract template instantiate --profile contract --input-file template-instance.json",
			"contract-cli contract template instantiate --profile contract --as bot --data '{\"template_number\":\"TMP001\",\"create_user_id\":\"ou_xxx\"}'",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/template_instances",
			"bot: /open-apis/contract/v1/template_instances",
			"bot 创建模板实例时，请调用方自行在 JSON body 中提供 create_user_id。",
		},
	}
	registry["contract enum"] = helpTopic{
		Name:  "contract enum",
		Usage: []string{"contract-cli contract enum <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli contract enum list --type <enum-type> [flags]", "查询枚举值"},
		},
	}
	registry["contract enum list"] = helpTopic{
		Name:    "contract enum list",
		Summary: "查询枚举值。",
		Usage:   []string{"contract-cli contract enum list --type <enum-type> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--type <enum-type>", "必填，枚举类型"},
		}),
		Examples: []string{
			"contract-cli contract enum list --profile contract --type contract_status",
		},
		Notes: []string{
			"仅支持 --as user；走 /open-apis/contract/v1/mcp/enum_values。",
		},
	}
}

func addMDMHelp(registry map[string]helpTopic) {
	registry["mdm"] = helpTopic{
		Name:  "mdm",
		Usage: []string{"contract-cli mdm <resource> <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli mdm vendor <subcommand> [flags]", "交易方主数据"},
			{"contract-cli mdm legal <subcommand> [flags]", "法人主体主数据"},
			{"contract-cli mdm fields list [flags]", "字段配置查询"},
		},
	}
	registry["mdm vendor"] = helpTopic{
		Name:  "mdm vendor",
		Usage: []string{"contract-cli mdm vendor <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli mdm vendor list [flags]", "查询交易方列表"},
			{"contract-cli mdm vendor get <vendor-id> [flags]", "查询交易方详情"},
		},
	}
	registry["mdm vendor list"] = helpTopic{
		Name:    "mdm vendor list",
		Summary: "查询交易方列表。",
		Usage:   []string{"contract-cli mdm vendor list [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), listQueryFlags()),
		Examples: []string{
			"contract-cli mdm vendor list --profile contract --name 供应商 --page-size 10",
			"contract-cli mdm vendor list --profile contract --as bot --name V00000001 --user-id-type employee_id",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/vendors",
			"bot: /open-apis/mdm/v1/vendors",
			"--name 会映射到底层 query vendor。",
		},
	}
	registry["mdm vendor get"] = helpTopic{
		Name:    "mdm vendor get",
		Summary: "查询交易方详情。",
		Usage:   []string{"contract-cli mdm vendor get <vendor-id> [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli mdm vendor get <vendor-id> --profile contract",
			"contract-cli mdm vendor get <vendor-id> --profile contract --as bot --user-id-type employee_id",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/vendors/{vendor_id}",
			"bot: /open-apis/mdm/v1/vendors/{vendor_id}",
		},
	}
	registry["mdm legal"] = helpTopic{
		Name:  "mdm legal",
		Usage: []string{"contract-cli mdm legal <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli mdm legal list [flags]", "查询法人主体列表"},
			{"contract-cli mdm legal get <legal-entity-id> [flags]", "查询法人主体详情"},
		},
	}
	registry["mdm legal list"] = helpTopic{
		Name:    "mdm legal list",
		Summary: "查询法人主体列表。",
		Usage:   []string{"contract-cli mdm legal list [flags]"},
		Flags:   concatHelpFlags(openPlatformCommonFlags(), listQueryFlags()),
		Examples: []string{
			"contract-cli mdm legal list --profile contract --name 主体A --page-size 10",
			"contract-cli mdm legal list --profile contract --as bot --user-id-type employee_id",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/legal_entities",
			"bot: /open-apis/mdm/v1/legal_entities/list_all",
			"--name 会映射到底层 query legalEntity。",
		},
	}
	registry["mdm legal get"] = helpTopic{
		Name:    "mdm legal get",
		Summary: "查询法人主体详情。",
		Usage:   []string{"contract-cli mdm legal get <legal-entity-id> [flags]"},
		Flags:   openPlatformCommonFlags(),
		Examples: []string{
			"contract-cli mdm legal get <legal-entity-id> --profile contract",
			"contract-cli mdm legal get <legal-entity-id> --profile contract --as bot --user-id-type employee_id",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/legal_entities/{legal_entity_id}",
			"bot: /open-apis/mdm/v1/legal_entities/{legal_entity_id}",
			"bot 路由会额外透传同名 query legal_entity_id。",
		},
	}
	registry["mdm fields"] = helpTopic{
		Name:  "mdm fields",
		Usage: []string{"contract-cli mdm fields <subcommand> [flags]"},
		Commands: []helpCommand{
			{"contract-cli mdm fields list --biz-line <biz-line> [flags]", "查询字段配置"},
		},
	}
	registry["mdm fields list"] = helpTopic{
		Name:    "mdm fields list",
		Summary: "查询字段配置。",
		Usage:   []string{"contract-cli mdm fields list --biz-line <biz-line> [flags]"},
		Flags: concatHelpFlags(openPlatformCommonFlags(), []helpFlag{
			{"--biz-line <biz-line>", "必填；user 支持 vendor、legal_entity、vendor_risk；bot 支持 vendor、legalEntity，legal_entity 会自动映射为 legalEntity"},
		}),
		Examples: []string{
			"contract-cli mdm fields list --profile contract --biz-line vendor",
			"contract-cli mdm fields list --profile contract --as bot --biz-line legal_entity",
			"contract-cli mdm fields list --profile contract --as bot --biz-line vendor --user-id-type employee_id",
		},
		Notes: []string{
			"user: /open-apis/contract/v1/mcp/config/config_list",
			"bot: /open-apis/mdm/v1/config/config_list",
			"bot 后端当前只接受 vendor 或 legalEntity；CLI 会把 bot 下的 legal_entity 映射为 legalEntity，vendor_risk 不支持 bot。",
		},
	}
}

func openPlatformCommonFlags() []helpFlag {
	return []helpFlag{
		{"--profile <name>", "profile 名称；不传使用当前 profile"},
		{"--as <user|bot>", "请求身份；不传使用 profile default_identity"},
		{"--output <json|yaml|table>", "输出格式，默认 json"},
		{"--raw", "原样输出响应 body"},
		{"--user-id-type <type>", "通用 query 参数 user_id_type；不传默认 user_id，传了则覆盖默认值"},
		{"--user-id <id>", "通用 query 参数 user_id；传了就透传，不传就不带"},
	}
}

func jsonBodyFlags() []helpFlag {
	return []helpFlag{
		{"--input-file <path>", "从 JSON 文件读取请求体；与 --data 互斥"},
		{"--data <json>", "内联 JSON 请求体；与 --input-file 互斥"},
	}
}

func pageFlags() []helpFlag {
	return []helpFlag{
		{"--page-size <n>", "分页大小"},
		{"--page-token <token>", "分页 token"},
	}
}

func listQueryFlags() []helpFlag {
	return concatHelpFlags([]helpFlag{
		{"--name <name>", "名称或编码查询条件"},
	}, pageFlags())
}

func concatHelpFlags(groups ...[]helpFlag) []helpFlag {
	var total int
	for _, group := range groups {
		total += len(group)
	}
	flags := make([]helpFlag, 0, total)
	for _, group := range groups {
		flags = append(flags, group...)
	}
	return flags
}
