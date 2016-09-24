// This is a command line wrapper for shell/bash commands that will
// Send a notification when command execution is done so that you
// can fire any long command and do whatever you want then get notified
// when command execution is done via HipChat and Slack.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bluele/slack"
	"github.com/caarlos0/env"
	"github.com/codeskyblue/go-sh"
	"github.com/tbruyelle/hipchat-go/hipchat"
	"github.com/urfave/cli"
)

var (
	logger *log.Logger
)

// notificator is a struct that holds needed information to send notifications.
type notificator struct {
	HipChatAuthToken string `env:"HIPCHAT_AUTHTOKEN"` // HipChat API Authentication Token.
	HipChatUser      string `env:"HIPCHAT_USER"`      // HipChatUser to be notified.
	SlackAuthToken   string `env:"SLACK_AUTHTOKEN"`   // Slack API Authentication Token.
	SlackUser        string `env:"SLACK_USER"`        // SlackUser to be notified.
}

func init() {
	// set log file to user homedir/.saydone.log
	logFile, err := os.OpenFile(filepath.Join(os.Getenv("HOME"), ".saydone.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("init: can not open logfile %s/saydone.log\n%v", os.Getenv("home"), err)
	}
	logger = log.New(logFile, "saydone: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	// Creating a new CLI App as a command wrapper
	app := cli.NewApp()
	app.Name = "saydone"
	app.Usage = "Runs a shell script and notify back when it's done"
	app.Version = "0.1"
	app.Author = "Ahmed Kamel"
	app.Action = appAction
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

// appAction is the main function that gets executed when saydone binary is invoked.
func appAction(c *cli.Context) error {
	// Show help if no Args passed.
	if c.NArg() == 0 {
		err := cli.ShowAppHelp(c)
		if err != nil {
			panic(err)
		}
		return nil
	}

	// Storing exported hipchat environment variables into n.
	var n notificator
	err := env.Parse(&n)
	if err != nil {
		logger.Println("Cannot parse env vars", err)
	}

	// Check if all required env vars are not set and stop program execution if they are.
	if n.HipChatAuthToken == "" && n.SlackAuthToken == "" {
		fmt.Println("Environment vars are not set, Please set at least one endpoint vars for the application to work.")
		fmt.Println("Exiting...")
		os.Exit(1)
	}

	// Run your command line arguments after saydone and
	// store output into cmdOutput variable.
	cmdOutput, err := sh.Command(c.Args().First(), c.Args().Tail()).SetTimeout(time.Hour * 24).CombinedOutput()
	if err != nil {
		logger.Println(err)
		return err
	}

	// Getting server hostname to be send as a message prefix
	serverHostname, err := os.Hostname()
	if err != nil {
		logger.Println("Error fetching hostname")
	}

	// Message to be sent as a notification
	notificationMsg := fmt.Sprintf(`Command running at %s is done with the following output:
%s`, serverHostname, string(cmdOutput))

	// Notify hipchat
	err = n.hipchat(notificationMsg)
	if err != nil {
		fmt.Println(err)
		logger.Println(err)
	}

	// Notify slack
	err = n.slack(notificationMsg)
	if err != nil {
		fmt.Println(err)
		logger.Println(err)
	}

	// Finally sending command output to terminal stdout.
	fmt.Fprintf(os.Stdout, string(cmdOutput))
	return nil
}

// slack is a method that takes message string and sends a slack
// notification message with executed command output to user specified in env var.
func (n notificator) slack(msg string) error {
	// Send slack notification if environment variables are set.
	api := slack.New(n.SlackAuthToken)
	err := api.ChatPostMessage(fmt.Sprintf("@%s", n.SlackUser), msg, nil)
	if err != nil {
		logger.Println(err)
		return errors.New("Failed to send msg to slack, Please check your env vars.")
	}
	return nil
}

// hipchat is a method that takes message string and sends a hipchat
// notification message with executed command output to user specified in env var.
func (n notificator) hipchat(msg string) error {
	// Send a HipChat notification if environment variables for credentials
	// are set.
	// Create a hipchat client to send notifications.
	hcClient := hipchat.NewClient(n.HipChatAuthToken)

	// Message to be sent to hipchat user.
	msgRequest := hipchat.MessageRequest{
		Message:       msg,
		Notify:        true,
		MessageFormat: "text"}

	// Messaging hipchat user with command CombinedOutput.
	_, err := hcClient.User.Message(n.HipChatUser, &msgRequest)
	if err != nil {
		logger.Println(err)
		return errors.New("Failed to send msg to hipchat, Please check your env vars.")
	}
	return nil
}
