pushd dist
for D in ../cmd/*
do
	export GOARCH=amd64
	export GOOS=windows
	go build -o $(basename $D)-${GOOS}-${GOARCH}.exe $D
	export GOOS=linux
	go build -o $(basename $D)-${GOOS}-${GOARCH} $D
	export GOOS=darwin
	go build -o $(basename $D)-${GOOS}-${GOARCH} $D
done
popd