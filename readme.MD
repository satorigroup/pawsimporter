PAWS Custom Fields Importer
==========


## Development

install Go 1.6 language compiler such as apt-get or yum

```sh
$ brew install go

You need to set up your [GOPATH](http://golang.org/doc/code.html#GOPATH).
For example, you can put the following lines in ~/.bash_profile or ~/.zshrc depends on the shell type
```
export GOPATH="$HOME/gocode"
export PATH=$PATH:$GOPATH/bin
```
create the following folder structure
```sh
cd ~
mkdir -p ~/gocode/src/satori/pawsimporter
```
so the folder structure would look like ~/gocode/src/satori/pawsimporter

Now cd into pawsimporter folder and clone the repo
```sh
cd ~/gocode/src/satori/pawsimporter
git clone https://github.com/satorigroup/pawsimporter
```

Update pawsimporter/.env file for local development:
```
MSSQL_CONN_STRING=server=172.23.0.22;user id=sa;password=password;port=1433;database=PAWS
DATA_FILE=data/data.xlsx
RUN_MODE=warning or insert
READ_ROWS=1-1000 or 2000
```


### Build, Import and Start

```sh
$ cp .env.example .env # change it match your local setting
$ ./script/build # to build the app
$ ./script/start # to start the app
```