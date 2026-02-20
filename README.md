# Cierge - Your restaurant reservation concierge

After too many times spent getting ready at a set time to then see all reservations get sniped by bots and disappear before I can click any, I decided to build something that would allow my friends and I to actually be able to get reservations and not have to organise a morning around getting online at a set time.  
**This is solely intended for personal use.** To all those who scalp and resell, you truly ruin it for everyone else and I hate you. 

## Introduction

Cierge is a platform built to support multiple users and handle scheduling reservation jobs. The goal is for a user to select the restaurant, the date, and time slots and Cierge will handle scheduling the execution of the reservation job right when reservations open up for the given date, executing it, and notifying the user of the result.  

Cierge is designed to be cloud agnostic but has native support for AWS to allow for the use of Lambdas to perform the job execution and EventBridge to handle scheduling.

The repository is a monorepo and contains multiple components that are each individually versioned. See [Components](#components)

> [!WARNING]
> Cierge is in active development. Please make sure to use a tagged version of the repo for actual usage.

## Getting Started

### I am a user
[![Latest CLI Release](https://img.shields.io/github/v/tag/daylamtayari/Cierge?filter=cli%2F*&label=CLI%20Release)](https://github.com/daylamtayari/Cierge/releases?q=cli)

1. Download the binary for the latest release of the command line interface that corresponds to your platform (Mac users, select Darwin)
    - Open a terminal and run the Cierge CLI by doing `./cierge`
2. Run `cierge init` and specify the server host provided to you by the server administrator and your credentials
3. Connect reservation platforms by running `cierge token add`
4. Verify everything is good by running `cierge status`

You are now ready to start creating reservation jobs!

To create a new reservation job, run `cierge job create` and follow the prompts.

### I want to host my server

1. Create the necessary AWS infrastructure (lambda, KMS key, roles) using the `deploy/aws.tf` Terraform  
    - A local implementation is on the roadmap for the next release but at this time, AWS usage is required
2. Generate TLS certificates for the host (required in production)
2. Complete the server configuration file (`deploy/server.json`)
    - Complete the AWS configuration using the values outputted from the Terraform
    - Generate a JWT secret (recommend `openssl rand -base64 64`)
    - Generate and set the database password
3. Run the Docker compose file that corresponds to your desired environment

Enjoy!


## Features
- Automated reservation booking at the exact time reservations become available
- Drop configurations so you don't have to manually set when the reservation needs to be executed
- Handles multiple acceptable slot times with respective priority to try and ensure that preferred times are booked
- Management of token platforms and maintaining of token lifecycle to allow for seamless experience for users
- Cloud agnostic job execution
- Command line interface for interacting with Cierge
- API library to allow for easy integrations

## Components

| Component | Description | |
|-----------|-------------|---|
| [api](https://github.com/daylamtayari/Cierge/tree/main/api) | API library for the Cierge API | [![Go Reference](https://pkg.go.dev/badge/github.com/daylamtayari/cierge/api.svg)](https://pkg.go.dev/github.com/daylamtayari/cierge/api) |
| [cli](https://github.com/daylamtayari/Cierge/tree/main/cli) | Command line interface for Cierge | |
| [deploy](https://github.com/daylamtayari/Cierge/tree/main/deploy) | Infrastructure as code for server deployment | |
| [errcol](https://github.com/daylamtayari/Cierge/tree/main/errcol) | Error collector designed for wide event logging | [![Go Reference](https://pkg.go.dev/badge/github.com/daylamtayari/cierge/errcol.svg)](https://pkg.go.dev/github.com/daylamtayari/cierge/errcol) |
| [lambda](https://github.com/daylamtayari/Cierge/tree/main/lambda) | AWS Lambda for the reservation job execution | |
| [opentable](https://github.com/daylamtayari/Cierge/tree/main/opentable) | OpenTable API library | [![Go Reference](https://pkg.go.dev/badge/github.com/daylamtayari/cierge/opentable.svg)](https://pkg.go.dev/github.com/daylamtayari/cierge/opentable) |
| [querycol](https://github.com/daylamtayari/Cierge/tree/main/querycol) | Database query collector designed for wide event logging | [![Go Reference](https://pkg.go.dev/badge/github.com/daylamtayari/cierge/querycol.svg)](https://pkg.go.dev/github.com/daylamtayari/cierge/querycol) |
| [reservation](https://github.com/daylamtayari/Cierge/tree/main/reservation) | Reservation job execution logic | [![Go Reference](https://pkg.go.dev/badge/github.com/daylamtayari/cierge/reservation.svg)](https://pkg.go.dev/github.com/daylamtayari/cierge/reservation) |
| [resy](https://github.com/daylamtayari/Cierge/tree/main/resy) | Resy API library | [![Go Reference](https://pkg.go.dev/badge/github.com/daylamtayari/cierge/resy.svg)](https://pkg.go.dev/github.com/daylamtayari/cierge/resy) |
| [server](https://github.com/daylamtayari/Cierge/tree/main/server) | Cierge server | |

## Roadmap

### Alpha (we are here)

Core functionality and good enough to start sharing.

### Beta (in progress)

- [ ] Notifications
- [ ] OpenTable support
- [ ] Local 'cloud' implementation
- [ ] Platform token lifecycle management
- [ ] Complete user management functionality
- [ ] Favourite restaurant functionality
- [ ] Complete command line interface

### Release

- [ ] Website (ugh)
- [ ] Complete documentation
- [ ] OIDC authentication
- [ ] Social authentication
