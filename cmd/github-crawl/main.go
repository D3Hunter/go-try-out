package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-lark/lark"
	"github.com/go-lark/lark/card"
	"github.com/google/go-github/v67/github"
	"gopkg.in/yaml.v3"
	"try-out/pkg/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("failed to run", err)
		os.Exit(1)
	}
}

func run() error {
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.Parse()

	cfg, err := parseConfig(configFile)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	formatedStartDate := now.AddDate(0, 0, -14).Format("2006-01-02")

	qualifiers := make([]string, 6+len(cfg.Repos)+len(cfg.Users))
	qualifiers = append(qualifiers, "is:pr")
	qualifiers = append(qualifiers, "is:open")
	qualifiers = append(qualifiers, "is:open")
	qualifiers = append(qualifiers, "-label:do-not-merge/needs-linked-issue")
	qualifiers = append(qualifiers, "-label:do-not-merge/work-in-progress")
	qualifiers = append(qualifiers, fmt.Sprintf("created:>=%s", formatedStartDate))
	for _, r := range cfg.Repos {
		qualifiers = append(qualifiers, fmt.Sprintf("repo:%s", r))
	}
	for u := range cfg.Users {
		qualifiers = append(qualifiers, fmt.Sprintf("author:%s", u))
	}

	qStr := strings.Join(qualifiers, " ")

	fmt.Printf("querying str: %s\n", qStr)

	ctx := context.Background()
	client := github.NewClient(nil)
	issues, _, err := client.Search.Issues(ctx, qStr, &github.SearchOptions{Sort: "created", Order: "asc"})
	if err != nil {
		return err
	}

	var markdown string
	for _, issue := range issues.Issues {
		if strings.Contains(strings.ToLower(*issue.Title), "wip") {
			fmt.Println("skip WIP PR: ", *issue.Title, *issue.HTMLURL)
			continue
		}
		additionalInfo := ""
		for _, label := range issue.Labels {
			if *label.Name == "needs-1-more-lgtm" {
				additionalInfo = "LGTM-1 ✅ "
				break
			}
		}
		timePart := ""
		days := int(time.Since(issue.CreatedAt.Time).Round(24*time.Hour).Hours() / 24)
		if days >= 3 {
			timePart = fmt.Sprintf("<font color='red'> created **%d** days ago.</font>", days)
		} else if days > 0 {
			timePart = fmt.Sprintf(" created **%d** days ago 📢.", days)
		}
		markdown += fmt.Sprintf("- %s[%s](%s). @%s%s\n",
			additionalInfo, *issue.Title, *issue.HTMLURL,
			*issue.User.Login, timePart)
	}
	cardBuilder := lark.NewCardBuilder()
	card := cardBuilder.Card(
		card.Markdown(markdown),
	).Blue().Title("PRs need to be reviewed (created in recent 2 weeks)")
	message := lark.NewMsgBuffer(lark.MsgInteractive).Card(card.String()).Build()
	bot := lark.NewNotificationBot(cfg.BotAddr)
	_, err = bot.PostNotificationV2(message)

	return err
}

func testSend() error {
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.Parse()

	cfg, err := parseConfig(configFile)
	if err != nil {
		return err
	}
	b := lark.NewCardBuilder()
	card := b.Card(
		b.Div(
			b.Field(b.Text("左侧内容PRs need to be review")).Short(),
			b.Field(b.Text("右侧内容")).Short(),
			b.Field(b.Text("整排内容")),
			b.Field(b.Text("整排**Markdown**内容").LarkMd()),
		),
		b.Div().
			Text(b.Text("Text Content")).
			Extra(b.Img("img_a7c6aa35-382a-48ad-839d-d0182a69b4dg")),
		b.Note().
			AddText(b.Text("Note **Text**").LarkMd()).
			AddImage(b.Img("img_a7c6aa35-382a-48ad-839d-d0182a69b4dg")),
	).Wathet().Title("Notification Card")
	message := lark.NewMsgBuffer(lark.MsgInteractive).Card(card.String()).Build()
	bot := lark.NewNotificationBot(cfg.BotAddr)
	resp, err := bot.PostNotificationV2(message)
	_ = resp
	return err
}

func parseConfig(fileName string) (*config.GithubCrawlConfig, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	cfg := &config.GithubCrawlConfig{}
	if err = decoder.Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
