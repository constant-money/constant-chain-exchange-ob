echo "Start build Constant Exchange Orderbook Service"

git pull

echo "Package install"
dep ensure -v

APP_NAME="constant-exchange-ob"

echo "go build -o $APP_NAME"
go build -o $APP_NAME

echo "cp ./$APP_NAME $GOPATH/bin/$APP_NAME"
mv ./$APP_NAME $GOPATH/bin/$APP_NAME

echo "Build Constant Exchange Orderbook success!"

export GOOGLE_APPLICATION_CREDENTIALS=~/Downloads/cash-prototype-52c1bc3c94d0.json