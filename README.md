# discordgoext

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)
[![Patreon](https://img.shields.io/badge/patreon-donate!-orange.svg?style=flat-square)](https://www.patreon.com/floretta)

## Deprecation notice

It's been two years since I have worked on this project. I'm not even sure if it still works. Given the new additions to the Discord API and future bots' reliance on them for their functionality, I have created a new and improved project called [Harmonia](https://github.com/Moonlington/harmonia). Please use that module instead.

## Current features

- Easier way of creating commands for a bot
- Automatic help command

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Contribute](#contribute)
- [License](#license)

## Install

```bash
go get -u github.com/Moonlington/discordgoext
```

## Usage

Here's an example of a bot that pings and pongs;

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/Moonlington/discordgoext"
)

func main() {

    // Create a new Discord session using the provided bot token.
    sess, err := discordgoext.New("Bot " + "INSERT YOUR BOT TOKEN", "pingbot.", false)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return
    }

    sess.AddCommand("Pings", discordgoext.NewCommand("ping", "Pings", "", "", func(ctx *discordgoext.Context) {
        ctx.SendMessage("pong!")
    }))
    sess.AddCommand("Pongs", discordgoext.NewCommand("pong", "Pongs", "", "", func(ctx *discordgoext.Context) {
        ctx.SendMessage("ping!")
    }))

    sess.AddCommand("Saying Hi", discordgoext.NewCommand(
        "hi",
        "Says hi",
        "[to who]",
        "If `to who` is not provided, it will say hi to you.",
        func(ctx *discordgoext.Context) {
            if len(ctx.Argstr) != 0 {
                ctx.SendMessage("Hi " + ctx.Argstr + "!")
            } else {
                ctx.SendMessage("Hi " + ctx.Mess.Author.Username + "!")
            }
        },
    ))

    // Open a websocket connection to Discord and begin listening.
    err = sess.Open()
    if err != nil {
        fmt.Println("error opening connection,", err)
        return
    }

    // Wait here until CTRL-C or other term signal is received.
    fmt.Println("Bot is now running.  Press CTRL-C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    // Cleanly close down the Discord session.
    sess.Close()
}
```

## Contribute

PRs accepted.

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

MIT © Moonlington
