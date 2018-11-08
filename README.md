# deliverbot
## Demo
| Demo |
|:-:|
|![jul-29-2018 19-26-14](https://user-images.githubusercontent.com/38090650/43365311-e3b342ac-9365-11e8-925a-44366cbb876e.gif)|


## Install
```bash
$ go get -u github.com/kishikawakatsumi/deliverbot
```

## Usage
### Basic
After install `deliverbot`, `deliverbot` execution binary is created under `$GOPATH/bin`. And it can used.
```
$ $GOPATH/bin/deliverbot --config ./config.toml
```

### Advanced
`deliverbot` can specify other options.
You can get these options from `deliverbot --help`.

```
 --config value, -c value  Load configuration *.toml
 --port value, -p value    Server port to be listened (default: "3000")
 --region value, -r value  Setting AWS region for tomlssm (default: "ap-northeast-1")
 --help, -h                show help
 --version, -v             print the version
```

## Development
If you want to develop `deliverbot` based original bot, it proceed with development in the following procedure.

1. `$ cd $DELIVERBOT_INSTALL_DIRECTORY`.
Probably it will be under the $GOPATH/src/github.com/kishikawakatsumi/deliverbot/, If `deliverbot` installed to use `go get`.
```bash
$ cd $GOPATH/src/github.com/kishikawakatsumi/deliverbot/
```

2. It necessary install dependencies.
```bash
$ make dep
```

3. You start to edit awesome customization for `deliverbot`.
```bash
$ vim .
```

4. You should make execution binary for local development. And after `make build`, you can get execution binary file for deliverbot. 
```bash
$ make build
$ ls bin | grep deliverbot
deliverbot # You get awesome deliverbot.
```

5. You can conform to execute your awesome deliverbot.
```bash
$ ./bin/deliverbot
```

It is okay to repeat from 3 to 5 steps. Enjoy development!
