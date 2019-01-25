sudo apt-get purge golang*
#sudo apt-get install golang-go golang-go.tools

sudo mkdir /opt/golang
sudo chmod 777 /opt/golang
cd /opt/golang
curl -O https://storage.googleapis.com/golang/go1.11.2.linux-amd64.tar.gz
tar -xvf go1.11.2.linux-amd64.tar.gz
sudo mv go /usr/local


export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin

cd /opt/anykey
git clone https://github.com/mixamarciv/admindb_app.git
cd admindb_app

export GOPATH=`pwd`
export PATH=$PATH:$GOPATH


go get "github.com/gorilla/sessions"
go get "github.com/gorilla/context"
go get "github.com/gorilla/mux"
go get "github.com/mixamarciv/gofncstd3000"
go get "github.com/go-gomail/gomail"
go get "github.com/nakagami/firebirdsql"
go get "github.com/nfnt/resize"
go get "github.com/tdewolff/minify"
go get "github.com/davecgh/go-spew/spew"
go get "github.com/qiniu/iconv"

go install
go build





