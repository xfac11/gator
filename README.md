# gator 
Browse your favorite rss feeds.

Supports Linux Ubuntu. Althought it can be used on mac or windows the commands and package manager need to be adapted.

## Prerequisites
- PostgreSQL
  - Install using your package manager. ``apt install postgresql``
  - Setup your postgres server and start it.
- Go
  - Install using Go [installation docs](https://go.dev/doc/install)
- goose
  - Install with: ``go install github.com/pressly/goose/v3/cmd/goose@latest``

## Tested On
Linux (Ubuntu)

## Installation
Use ``go install`` inside the root of the repository to install the gator CLI. 

This will install the gator CLI in the directory named by the GOBIN environment variable, which defaults to $GOPATH/bin or $HOME/go/bin if the GOPATH environment variable is not set. Read the [documentation](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies) about go install for more informaton.

Create one database called gator with postgres and then use goose inside sql/schema to migrate to the latest version.

``goose postgres <db_connection_string> up``

## Config file
Gator uses one config file to connect to the database and keep track of the current user.
- Create a file in your current users' home directory called ``.gatorconfig.json``
- Paste this (The user will be inserted by gator) `` {"db_url":"postgres://postgres:postgres@localhost:5432/gator"} ``
- The url structure is: ``"postgres://username:password@localhost:5432/gator``. Driver is always postgres and username and password can be different if you have setup PostgreSQL differently. localhost is the host name and 5432 is the port. Use the port where you run PostgreSQL which is usually 5432.

## Running the CLI
With the previous steps done you are ready to run gator.
To run simply do in your terminal ``gator [command] [arguments]`` Some example commands are:
- register "user_name"
  - Adds one user with that name
- login "user_name"
  - Changes the current user
- addfeed "feed_name" "feed_url"
  - Adds one feed with that name
- follow "feed_name"
  - Adds the feed to the current users' follow list
- users
  - Outputs all registered users and the current user
- agg "time_between_requests"
  - This is supposed to be run in the background but can be used to run for some time and then ctrl-c out from it. It fetches posts and inserts them in the database
- browse "limit"
  - Browse "limit" amount of posts that are fetched using agg