Je participe ("I participate")
============

"Je participe" is a doodle-like service to organize an event with many activities and participants.
This service was initially built to organize end of the year party for schools. 

An activity is a task or responsability during the event like running the bar from 14h00 to 16h00, making some cookies, housecleening on sunday, ...

The purpose of the project is to simplify the task for event organizers to get volonteers and know who accept to do what and when.

  * First, the organizer describes all the activities in the event board page
  * Then, volonteers write their names in every activity they like (For real, organizers send some stimulant email to motivate them)
  * Step by step, all the activities gets volunteers and reach the maximum participants limit (if there is one in your case, free to you).
  * To finish, organizer closes the inscription for each activity. If needed, very rare in real cases, organiser can remove people from an activity.

Main features :
  * Each event can have multiple activities (One event = many doodles in one page)
  * Each activity has its own life cycle, list of participants (or volonteers)
  * No account is needed for a participant to volountrer to an activity or access the event board (so everybody can write other people names without troubles. This is important because in typical situation, people volunteer as a group, only the responsible of the group writes down the names on the board. The service is based on trust.)
  * Each participant can send public information (like their names) and private information (like their phone number) when they volonteer.
  * Private information are only visible by the organizer and the volunteer itself
  * If a volunteer wants to cancel its participation, he can if he is on same computer (same IP). Else he has to ask the organizer by email for that. Perhaps, this is not obvious, but this rules works fine in previous events without any claim (more than 50 volunteers).

## Project Status

Project is in active developpement.

Database support is limited to [BoltDB -- an embedded key/value database for Go](https://raw.githubusercontent.com/boltdb)

Sending email is limited to [Mailjet](https://mailjet.com/)

## Getting Started

### Installing

Create a mailjet account and add your API keys to your path :

```sh
export MJ_APIKEY_PUBLIC=xxx
export MJ_APIKEY_PRIVATE=xxx
```

Run a server from code (to be executed in your GOPATH) :

```sh
git clone https://github.com/julienbayle/jeparticipe
go get ./...
go run cmd/main.go
```

### Quick project description

app : The application

cmd : Main GO File

email : Sending email tools

entities : Object models

services : REST API methods implementation

templates : mail templates

### API Description

To be added

## ROAD MAP

  * Add a report API
  * Send email to volunteers from the service
  * List all events
  * Support multi-languages
  * Video presentation
