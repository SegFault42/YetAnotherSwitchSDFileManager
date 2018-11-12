echo "Cron" >> /tmp/cron.log

cd /home/ubuntu/go/src/github.com/SegFault42/YetAnotherSwitchSDFileManager > /tmp/cron.log
go run main.go > /tmp/cron.log
git add . > /tmp/cron.log
git commit -m "update SDFile" > /tmp/cron.log
git push > /tmp/cron.log

