package katia

import "github.com/bwmarrin/discordgo"

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

	return component.OnExecute(ctx)
}
