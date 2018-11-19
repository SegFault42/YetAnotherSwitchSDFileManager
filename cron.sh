echo "Cron" >> /tmp/cron.log

export DEVKITPRO=/opt/devkitpro
export DEVKITARM=/opt/devkitpro/devkitARM
export DEVKITPPC=/opt/devkitpro/devkitPPC
export DEVKITA64=/opt/devkitpro/devkitA64
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/usr/local/go/bin


cd /home/ubuntu/go/src/github.com/SegFault42/YetAnotherSwitchSDFileManager
go run main.go
git add .
git commit -m "update SDFile"
git push

