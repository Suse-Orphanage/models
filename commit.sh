git add .
git commit -m "$1"
git tag $2-postgres
git push origin master $2-postgres