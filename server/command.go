package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func getAutoCompleteDesc(m map[string]commandHandlerFunc) string {
	var b strings.Builder

	b.WriteString("Available commands: ")

	var i int
	for k := range m {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k)
	}

	return b.String()
}

func getLogsAutocompleteData() *model.AutocompleteData {
	getLogs := model.NewAutocompleteData("get-logs", "[alerts|notifications]", "Get logs of specific type")

	alerts := model.NewAutocompleteData("alerts", "", "Get alerts")
	getLogs.AddCommand(alerts)

	notifications := model.NewAutocompleteData("notifications", "", "Get notifications")
	getLogs.AddCommand(notifications)

	return getLogs
}

func getAutocompleteData(desc string) *model.AutocompleteData {
	nagios := model.NewAutocompleteData("nagios", "[command]", desc)

	// Auto-complete for get-logs command.
	nagios.AddCommand(getLogsAutocompleteData())

	// Auto-complete for set-logs-limit command.
	setLogsLimit := model.NewAutocompleteData("set-logs-limit", "[positive integer]", "Set maximum number of logs to display")
	nagios.AddCommand(setLogsLimit)

	// Auto-complete for set-logs-start-time command.
	setLogsStartTime := model.NewAutocompleteData("set-logs-start-time", "[seconds]", "Set number of seconds to get logs from")
	nagios.AddCommand(setLogsStartTime)

	// Auto-complete for set-report-frequency command.
	setReportFrequency := model.NewAutocompleteData("set-report-frequency", "[minutes]", "Set frequency of system monitoring reports")
	nagios.AddCommand(setReportFrequency)

	return nagios
}

func (p *Plugin) getCommand() *model.Command {
	desc := getAutoCompleteDesc(p.commandHandlers)

	return &model.Command{
		Trigger:          "nagios",
		AutoComplete:     true,
		AutoCompleteDesc: desc,
		AutoCompleteHint: "[command]",
		DisplayName:      "Nagios",
		Description:      "A Mattermost plugin to interact with Nagios",
		AutocompleteData: getAutocompleteData(desc),
		// AutocompleteIconData: nil,
	}
}

func parseCommandArgs(args *model.CommandArgs) (command, action string, parameters []string) {
	fields := strings.Fields(args.Command)

	if len(fields) > 0 {
		command = fields[0]
	}
	if len(fields) > 1 {
		action = fields[1]
	}

	parameters = fields[2:]

	return command, action, parameters
}

func (p *Plugin) getCommandResponse(args *model.CommandArgs, text string) *model.CommandResponse {
	p.API.SendEphemeralPost(args.UserId, &model.Post{
		UserId:    p.botUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	})
	return &model.CommandResponse{}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	command, action, parameters := parseCommandArgs(args)

	if command != "/nagios" {
		return &model.CommandResponse{}, nil
	}

	var msg string

	if f, ok := p.commandHandlers[action]; ok {
		msg = f(p.API, p.client, parameters)
	} else {
		msg = fmt.Sprintf("Unknown action (%s).", action)
	}

	return p.getCommandResponse(args, msg), nil
}
