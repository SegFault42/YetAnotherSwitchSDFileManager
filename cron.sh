echo "Cron" >> /tmp/cron.log

export DEVKITPRO=/opt/devkitpro
export DEVKITARM=/opt/devkitpro/devkitARM
export DEVKITPPC=/opt/devkitpro/devkitPPC
export DEVKITA64=/opt/devkitpro/devkitA64
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/usr/local/go/bin


cd /home/ubuntu/go/src/github.com/SegFault42/YetAnotherSwitchSDFileManager > /tmp/cron.log
go run main.go > /tmp/cron.log
git add . > /tmp/cron.log
git commit -m "update SDFile" > /tmp/cron.log
git push > /tmp/cron.log

