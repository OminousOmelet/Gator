--GATOR--
- a blog aggregator created in Golang

--INSTALLATION (Linux)--
- Ensure that you have the latest version of Golang and Postgres installed
- in Postgres, create a new database called 'gator'
- In root, create a new json file named '.gatorconfig.json' and write in the following:
  {
    "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
    "current_user_name": ""
  }
- Replace 'username' and 'password' with your Postgres credentials. Leave current user blank.
- from root, create a 'gator' directory, then navigate to it
- type the command 'go install github.com/OminousOmelet/gator'

--COMMANDS--
'register'  - creates a new user
'login'     - set the current user
'addfeed'   - add a feed to the database
'follow'    - follow a specific feed in the database
'agg'       - aggregates the added feeds, posting their contents to the database
'browse'    - read contents of the posts
