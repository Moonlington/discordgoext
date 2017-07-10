# Discordflo

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)
[![Floretta's Coding Space](https://img.shields.io/badge/discord-Floretta's%20Coding%20Space-738bd7.svg?style=flat-square)](https://discordapp.com/invite/pPxa93F)
[![Patreon](https://img.shields.io/badge/patreon-donate!-orange.svg?style=flat-square)](https://www.patreon.com/floretta)

> It's the best, I think, Im not sure.

## Current features

Adding commands and having a cool looking help command.

## Table of Contents

-   [Install](#install)
-   [Usage](#usage)
-   [Contribute](#contribute)
-   [License](#license)

## Install

```
go get -u github.com/Moonlington/discordflo
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

	"github.com/Moonlington/discordflo"
)

func main() {

	// Create a new Discord session using the provided bot token.
	ffs, err := discordflo.New("Bot " + "INSERT YOUR BOT TOKEN", "pingbot.", false)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	ffs.AddCommand(discordflo.NewCommand("ping", "Pings", "", "", func(ctx *discordflo.Context) {
		ffs.ChannelMessageSend(ctx.Mess.ChannelID, "pong!")
	}), "Pings")
	ffs.AddCommand(discordflo.NewCommand("pong", "Pongs", "", "", func(ctx *discordflo.Context) {
		ffs.ChannelMessageSend(ctx.Mess.ChannelID, "ping!")
	}), "Pongs")

	// Open a websocket connection to Discord and begin listening.
	err = ffs.Open()
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
	ffs.Close()
}
```

## Contribute

PRs accepted.

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

MIT © Floretta
