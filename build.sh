#!/bin/bash
echo "Building feedinator"
pushd feed
go test
if [ $? -ne 0 ]
  then
	echo "Tests failed"
  else
	popd
    sh -c 'go build -o feedinator main.go'
fi

