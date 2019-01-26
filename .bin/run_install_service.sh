#!/usr/bin/env bash

file=/etc/systemd/system/site_anykey.service
echo "[Unit]" > $file
echo "Description=site_anykey" >> $file
echo "After=network.target" >> $file
echo " " >> $file
echo "[Service]" >> $file
echo "User=root" >> $file
echo "Type=simple" >> $file
echo "Restart=always" >> $file
echo "RestartSec=30" >> $file
echo "#UMask=022" >> $file
echo "ExecStart=/opt/anykey/admindb_app/.bin/run_app.sh" >> $file
echo " " >> $file
echo "[Install]" >> $file
echo "WantedBy=multi-user.target" >> $file


systemctl daemon-reload
#systemctl enable site_anykey - autostart service - автозапуск сервиса после старта/перезагрузки системы
systemctl enable site_anykey
systemctl start site_anykey
systemctl status site_anykey.service

