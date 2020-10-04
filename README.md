# Github Projects to Clubhouse Migration

This program is written to copy all of the current cards out of the selected Github Project and create them
 in Clubhouse as stories using both of their free APIs. This process will also move cards that have been converted
 to issues. Any cards/issues that contain markdown checkbox lists are converted into Clubhouse tasks.

To run the program you have to either grab the latest release or download and build the code for your platform.

This program is built to be interactive, so once you provde the [Clubhouse API](https://app.clubhouse.io/settings/account/api-tokens)
 and [GiHub API](https://github.com/settings/tokens) tokens the program it will guide you through the entire migration process.


## Usage:

```sh
ghch
```
> ```
> Commands:
>  migrate
>   Move all Github project cards to Clubhouse
>   -ch-token string
>         Clubhouse API Token
>   -gh-token string
>         Github API Token
> ```
```sh
ghch migrate -ch-token {clubhouse-token} -gh-token {github-token}
```
> Then follow the on-screen instructions


## Other Information

This program was built on top of Github's API **v3** (using [Google's Github library](https://github.com/google/go-github)) and
 the Clubhouse's API **v3** (using [a library forked that I've updated](https://github.com/nhalstead/clubhouse)) so its build
 to be compatible for a while till either company changes their API.
