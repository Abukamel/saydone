// This is a command line wrapper for shell/bash commands that will
// Send a notification when command execution is done so that you
// can fire any long command and do whatever you want then get notified
// when command execution is done via HipChat, Slack and Email.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bluele/slack"
	"github.com/caarlos0/env"
	"github.com/codeskyblue/go-sh"
	"github.com/tbruyelle/hipchat-go/hipchat"
	"github.com/urfave/cli"
)

// NotificationInfo is a struct that holds needed information to send notifications.
type NotificationInfo struct {
	HipChatAuthToken string `env:"HIPCHAT_AUTHTOKEN"` // HipChat API Authentication Token.
	HipChatUser      string `env:"HIPCHAT_USER"`      // HipChatUser to be notified.
	SayDoneEmail     string `env:"SAYDONE_EMAIL"`     // Email address to be notified.
	SlackAuthToken   string `env:"SLACK_AUTHTOKEN"`   // Slack API Authentication Token.
	SlackUser        string `env:"SLACK_USER"`        // SlackUser to be notified.
}

func main() {
	// Storing exported hipchat environment variables into nInfo
	var notificationInfo NotificationInfo
	err := env.Parse(&notificationInfo)
	if err != nil {
		log.Println(err)
	}

	// Creating a new CLI App as a command wrapper
	app := cli.NewApp()
	app.Name = "saydone"
	app.Usage = "Runs a shell script and notify back when it's done"
	app.Version = "0.1"
	app.Author = "Ahmed Kamel"
	app.Action = func(c *cli.Context) error {

		// Show help if no Args passed.
		if c.NArg() == 0 {
			err := cli.ShowAppHelp(c)
			if err != nil {
				panic(err)
			}
			return nil
		}

		// Run your command line arguments after saydone and
		// store output into out variable.
		out, err := sh.Command(c.Args().First(), c.Args().Tail()).CombinedOutput()
		if err != nil {
			log.Println(err)
			return err
		}

		// Getting server hostname to be send as a message prefix
		serverHostname, err := os.Hostname()
		if err != nil {
			log.Println("Error fetching hostname")
		}

		// Message to be sent as a notification
		notificationMsg := fmt.Sprintf(`Command running at %s is done with the following output:
%s`, serverHostname, string(out))

		// Send a HipChat notification if environment variables for credentials
		// are set.
		if notificationInfo.HipChatAuthToken != "" && notificationInfo.HipChatUser != "" {
			// Create a hipchat client to send notifications.
			hcClient := hipchat.NewClient(notificationInfo.HipChatAuthToken)

			// Message to be sent to hipchat user.
			msgRequest := hipchat.MessageRequest{
				Message:       notificationMsg,
				Notify:        true,
				MessageFormat: "text"}

			// Messaging hipchat user with command CombinedOutput.
			_, err = hcClient.User.Message(notificationInfo.HipChatUser, &msgRequest)
			if err != nil {
				log.Println(err)
			}
		} else {
			fmt.Println("Skipping hipchat notifications due to lake of information.")
			fmt.Println("Please check your environment variables.")
		}

		// Send an email notification if environment variable is set.
		if notificationInfo.SayDoneEmail != "" {
			// I have used sh command mail because of problems regarding certificate verification with golang net/smtp package.
			err := sh.Command("echo", notificationMsg).Command("mail", "-s", fmt.Sprintf("Command execution done at %s", serverHostname), notificationInfo.SayDoneEmail).Run()
			if err != nil {
				log.Println(err)
			}
		} else {
			fmt.Println("Skipping email notifications due to lake of information.")
			fmt.Println("Please check your environment variables.")
		}

		// Send slack notification if environment variables are set.
		if notificationInfo.SlackAuthToken != "" {
			api := slack.New(notificationInfo.SlackAuthToken)
			err := api.ChatPostMessage(fmt.Sprintf("@%s", notificationInfo.SlackUser), notificationMsg, nil)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Println("Skipping slack notifications due to lake of information.")
			fmt.Println("Please check your environment variables.")
		}
		// Finally sending command output to terminal stdout.
		fmt.Fprintf(os.Stdout, string(out))
		return nil
	}

	err = app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
