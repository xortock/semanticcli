#!/bin/bash

package=$1
binary_name=$2
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi
package_split=(${package//\// })
package_name=${package_split[-1]}
	
platforms=("windows/amd64" "windows/386" "linux/amd64" "linux/386" "linux/arm64" "linux/arm")

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=$binary_name'-'$GOOS'-'$GOARCH
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi	
	export GOOS=$GOOS 
    export GOARCH=$GOARCH 
    go build -o ./bin/$output_name -ldflags "-X main.version=$3" $package
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done