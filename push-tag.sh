export VERSION=`git rev-list --count HEAD`; 
echo Adding git tag with version v0.0.${VERSION};
git tag v0.0.${VERSION};
git push origin v0.0.${VERSION};
