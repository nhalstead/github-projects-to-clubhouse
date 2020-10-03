# Github to Clubhouse Migration

To run the program you have to either grab the latest release or download and build the code for your platform.

This program is built to be interactive, so once you provde the [Clubhouse API token](https://app.clubhouse.io/settings/account/api-tokens)
 and the [GiHub API token](https://github.com/settings/tokens).
 
 Once you provide them and execute the program it will walk you throught the entire migration process.

---

Command line params:

```
Commands:
 migrate
  Move all Github project cards to Clubhouse
  -ch-token string
        Clubhouse API Token
  -gh-token string
        Github API Token
```

---

This program was built on top of Github's API **v3** (using [Google's Github library](https://github.com/google/go-github)) and
 the Clubhouse's API **v3** (using [a library forked that I've updated](https://github.com/nhalstead/clubhouse)) so its build to be compatible for a while till either company changes their API.
