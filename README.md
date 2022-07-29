# Tsuki
Tsuki is a minimalistic open-sourced social media platform, built using Go.

## Running on local machine

### Requirements
- Tsuki requires a `PostgreSQL` database to store all the data.
- It uses the `Gmail API` for sending verification mail ([Reference](https://developers.google.com/gmail/api/quickstart/python)) and the `Freeimage API` for storing pictures ([Reference](https://freeimage.host/page/api)).
- It also requires some environment variables to be declared in the `.env` file. The variables can be found in `example.env`

### Installation
```
go mod download
```

### Building and Running
```
go build .
./tsuki-go
```
