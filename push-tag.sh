#!/bin/sh
export VERSION=`git rev-list --count HEAD`; 
echo Adding git tag with version v0.2.${VERSION};
git tag v0.2.${VERSION};
git push origin v0.2.${VERSION};
