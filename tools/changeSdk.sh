#!/usr/bin/env bash 
## NOTE: sed -i '' only works on mac, please remove the '' if you're on a *nix OS
REPO="github.com\/netsage-project\/sdk"
SDK="github.com\/grafana-tools\/sdk"

find . -iname "*.go" -exec sed -i '' -e "s/$REPO/$SDK/g" {} \;
echo sed -i '' -e "s/$REPO/$SDK/g" go.mod
echo "Updated all references to $(echo $SDK | sed -e 's/\\//'g ).  Please check your go.mod to update the version if points to. You may also wish to run go mod tidy as well"
