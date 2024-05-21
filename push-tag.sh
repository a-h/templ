#!/bin/sh
if [ `git rev-parse --abbrev-ref HEAD` != "main" ]; then
  echo "Error: Not on main branch. Please switch to main branch.";
  exit 1;
fi
git pull
if ! git diff --quiet; then
  echo "Error: Working directory is not clean. Please commit the changes first.";
  exit 1;
fi
export VERSION=`cat .version`
echo Adding git tag with version v${VERSION};
git tag v${VERSION};
git push origin v${VERSION};
