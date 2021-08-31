package katia

import "github.com/bwmarrin/discordgo"

func handleReady(bot *Bot, e *discordgo.Ready) {
	for _, command := range bot.commands {
		appID := bot.Session.State.User.ID
		app := &discordgo.ApplicationCommand{
			Name:        command.Name,
			Description: command.Description,
		}

		if command.Options != nil {
			app.Options = command.parseOptions(command.Options)
		}

		if _, err := bot.Session.ApplicationCommandCreate(appID, "872959811774459945", app); err != nil {
			bot.Logger.Error(err)
		}
	}
}

func handleInteractionCreate(bot *Bot, e *discordgo.InteractionCreate) {
	var result interface{}
	var executor string

	switch e.Type {
	case discordgo.InteractionApplicationCommand:
		result = handleInteractionApplicationCommand(bot, e)
		executor = e.ApplicationCommandData().Name
	case discordgo.InteractionMessageComponent:
		result = handleInteractionMessageComponent(bot, e)
		executor = e.MessageComponentData().CustomID
	}

	res := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	}

	switch r := result.(type) {
	case nil:
		return

	case error:
		bot.Logger.Error(r)
		res.Data.Embeds = []*discordgo.MessageEmbed{{
			Description: "내부 오류가 발생했습니다",
			Color:       0xFF0000,
		}}

	case bool:
		res.Type = discordgo.InteractionResponseDeferredMessageUpdate
	case string:
		res.Data.Content = r
	case discordgo.MessageEmbed:
		res.Data.Embeds = []*discordgo.MessageEmbed{&r}
	case []discordgo.MessageEmbed:
		res.Data.Embeds = []*discordgo.MessageEmbed{}
		for _, embed := range r {
			res.Data.Embeds = append(res.Data.Embeds, &embed)
		}
	case discordgo.InteractionResponse:
		res = &r
	case discordgo.InteractionResponseData:
		res.Data = &r
	default:
		bot.Logger.Warnf("%s 실행자가 지원하지 않는 %v 자료형을 반환했습니다: %+v", executor, r, r)
		res.Data.Content = `¯\_(ツ)_/¯`
	}

	bot.Session.InteractionRespond(e.Interaction, res)
}

func handleInteractionApplicationCommand(bot *Bot, e *discordgo.InteractionCreate) interface{} {
	data := e.ApplicationCommandData()
	command := bot.Command(data.Name)
	if command == nil {
		return nil
	}

	ctx := CommandContext{
		Bot:         bot,
		Command:     command,
		Data:        data,
		Interaction: e.Interaction,
	}

	if command.Options != nil {
		ctx.Options = ctx.formatOptions(command.Options, data.Options)
	}

	bot.Logger.Infof("%s requested '%s' command", e.Member.User.ID, data.Name)

	return command.OnExecute(ctx)
}

func handleInteractionMessageComponent(bot *Bot, e *discordgo.InteractionCreate) interface{} {
	data := e.MessageComponentData()
	component := bot.Component(data.CustomID)
	if component == nil {
		return nil
	}

	ctx := ComponentContext{
		Bot:         bot,
		Component:   component,
		Data:        data,
		Interaction: e.Interaction,
	}

	bot.Logger.Infof("%s requested '%s' component", e.Member.User.ID, data.CustomID)

	return component.OnExecute(ctx)
}
