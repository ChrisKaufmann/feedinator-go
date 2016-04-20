#!/bin/bash
echo "Testing feed"
pushd feed
go test
if [ $? -ne 0 ]
  then
	echo "Tests failed"
	failed=1
fi
popd

echo "Testing auth"
pushd auth
go test
if [ $? -ne 0 ]
  then
	echo "Tests failed"
	failed=1
fi

if [ $failed -ne 1 ]
	then	
		echo "Building feedinator"
		sh -c 'go build -o feedinator main.go'
fi
