## Description
This is a command wrapper that will notify you when a shell command execution
is done by sending a message with command output.
Currently it supports notfying via slack, hipchat and email.
## Usage
Export notification services environment variables.
`export HIPCHAT_AUTHTOKEN=your_hipchat_auth_token`
`export HIPCHAT_USER=hipchat_user_to_be_notified`
`export SLACK_AUTHTOKEN=your_slack_auth_token`
`export SLACK_USER=slack_user_to_be_notified`
`export SAYDONE_EMAIL=email_address_to_receive_notifications`
You can add global variables to your .bashrc file for persistency.

Command Usage: saydone yourCommand yourCommandOptions yourCommandArgs
e.g: saydone rsync -Pav /home/user/ /home/user2/
It will send a notification after command is done with command output combining stdout and stderr in notification message.

It will recognize the authentication to supported services via GLOBAL VARIABLES.
So you should export needed variables or put them in your /etc/profile, ~/.bash_profile,
~/.bashrc for permanent use or user export command for current session use e.g 
export "golabl_var=value" or you can prepend them before the command "saydone" for
a command only scope e.g global_var=value saydone ls -ltrha /home
