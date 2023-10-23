#!/bin/sh
export VERSION=`cat .version`
echo Adding git tag with version v${VERSION};
git tag v${VERSION};
git push origin v${VERSION};
