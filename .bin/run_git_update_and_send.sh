#!/usr/bin/env bash

cd ..

#git log - просмотр последних коммитов
#git reset --hard 9e7f2726f1902eb9bf3eda0d893221adb62e0eab - возврат к нужному коммиту с безвозвратным удалением изменений

DATE=`date '+%Y-%m-%d %H:%M:%S'`

git config --global user.email mixamarciv@gmail.com
git config --global user.name mixamarciv@gmail.com
git add .
git commit -a -m "changes ${DATE}"
git push
