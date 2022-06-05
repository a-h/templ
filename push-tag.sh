export VERSION=`git rev-list --count HEAD`; 
echo Adding git tag with version v2.0.${VERSION};
git tag v2.0.${VERSION};
git push origin v2.0.${VERSION};
