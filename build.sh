echo "Building feedinator"
sh -c 'go build -o feedinator main.go db.go auth.go util.go feed.go category.go entry.go'

echo "Building updater"
sh -c 'go build -o updater update.go db.go auth.go util.go feed.go category.go entry.go'
