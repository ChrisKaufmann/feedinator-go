#!/bin/bash
echo "Building feedinator"
pushd feed
go test
if [ $? -ne 0 ]
  then
	echo "Tests failed"
  else
    sh -c 'go build -o feedinator main.go'
fi
popd

