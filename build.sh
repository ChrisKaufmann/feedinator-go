echo "Building feedinator\n"
go build -o feedinator feedinator main.go db.go auth.go util.go feed.go category.go entry.go

echo "Building updater\n"
go build -o updater update.go db.go auth.go util.go feed.go category.go entry.go
