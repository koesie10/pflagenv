# pflagenv [![GoDoc](https://godoc.org/github.com/koesie10/pflagenv?status.svg)](https://godoc.org/github.com/koesie10/pflagenv)

## Install

```shell script
go get github.com/koesie10/pflagenv
```

## Examples

### Simple example

```go
var config = struct {
    Addr string `env:"ADDR" flag:"addr" desc:"the address for the server to listen on"`

    ChatAddr string `env:"CHAT_ADDR" flag:"chat-adr" desc:"the address of the chatsvc"`

    DiscordBotToken       string            `env:"DISCORD_BOT_TOKEN" flag:"discord-bot-token" desc:"the Discord bot token"`
    DiscordChatChannelIDs map[string]string `env:"DISCORD_CHAT_CHANNEL_IDS" flag:"discord-chat-channel-ids" desc:"the chat channel IDs in channelname=discordid format, case-sensitive"`
    DiscordChatName       string            `env:"DISCORD_CHAT_NAME" flag:"discord-chat-name" desc:"the chat name"`

    OAuthIssuer       string   `env:"OAUTH_ISSUER" flag:"oauth-issuer" desc:"OAuth issuer to retrieve configuration from"`
    OAuthClientID     string   `env:"OAUTH_CLIENT_ID" flag:"oauth-client-id" desc:"OAuth client ID"`
    OAuthClientSecret string   `env:"OAUTH_CLIENT_SECRET" flag:"oauth-client-secret" desc:"OAuth client secret"`
    OAuthScopes       []string `env:"OAUTH_SCOPES" flag:"oauth-scopes" desc:"OAuth scopes to request for the client-credentials flow"`
}{
    Addr: ":2561",

    ChatAddr: "127.0.0.1:2428",

    DiscordChatName: "Operations",

    OAuthIssuer: "http://localhost:4444",
}

flagSet := pflag.NewFlagSet("example", pflag.ExitOnError)

if err := pflagenv.Setup(flagSet, config); err != nil {
    log.Fatal(err)
}

if err := pflagenv.Parse(config); err != nil {
    log.Fatal(err)
}

fmt.Println(config.Addr)
```
