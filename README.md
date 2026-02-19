# Cierge - Your restaurant reservation concierge

After too many times spent getting ready on my laptop at a set time to then see all reservations get sniped by bots and disappear before I can click any, I decided to build something that would allow my friends and I to actually be able to get reservations and not have to organise a morning around getting online at a set time.  
**This is solely intended for personal use.** To all those who scalp and resell, you truly ruin it for everyone else.

## Introduction

Cierge is a platform built to support multiple users and handle scheduling reservation jobs. The goal is for a user to select the restaurant, the date, and time slots and Cierge will handle scheduling the execution of the reservation job right when reservations open up for the given date, executing it, and notifying the user of the result.  

Cierge is designed to be cloud agnostic but has native support for AWS to allow for using Lambdas to perform the execution and EventBridge to handle scheduling.

The repository is a monorepo and contains multiple components that are each individually versioned. See [Components](#components)

> [!WARNING]
> Cierge remains in active development. Please make sure to use a tagged version of the repo for actual usage.

## Getting Started

### I am a user

### I want to host my server

## Components

- [api](https://github.com/daylamtayari/Cierge/tree/main/api) - API library for the Cierge API
- [cli](https://github.com/daylamtayari/Cierge/tree/main/cli) - Command line interface for Cierge
- [deploy](https://github.com/daylamtayari/Cierge/tree/main/deploy) - Infrastructure as code for server deployment
- [errcol](https://github.com/daylamtayari/Cierge/tree/main/errcol) - Error collector designed for wide event logging
- [lambda](https://github.com/daylamtayari/Cierge/tree/main/lambda) - AWS Lambda for the reservation job execution
- [opentable](https://github.com/daylamtayari/Cierge/tree/main/opentable) - OpenTable API library
- [querycol](https://github.com/daylamtayari/Cierge/tree/main/querycol) - Database query collector designed for wide event logging
- [reservation](https://github.com/daylamtayari/Cierge/tree/main/reservation) - Reservation job execution logic
- [resy](https://github.com/daylamtayari/Cierge/tree/main/resy) - Resy API library
- [server](https://github.com/daylamtayari/Cierge/tree/main/server) - Cierge server

## Roadmap

### Alpha

Core functionality and good enough to start sharing.

### Beta

- [ ] Notifications
- [ ] OpenTable support
- [ ] Local 'cloud' implementation
- [ ] Platform token lifecycle management
- [ ] Complete user management functionality
- [ ] Favourite restaurant functionality
- [ ] Complete command line interface

### Release

- [ ] Website (ugh, 'web design is my passion')
- [ ] Complete documentation
- [ ] OIDC authentication
- [ ] Social authentication
