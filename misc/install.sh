service=prometheus-machinist-gateway

systemctl is-enabled $service
if [ $? -eq 0 ]; then
  systemctl stop $service
fi


cd ..
go build
install -Dm755 prometheus_machinist_gateway /usr/bin/prometheus_machinist_gateway
install -Dm644 example/prometheus_machinist_gateway.yml /etc/prometheus/prometheus_machinist_gateway.yml
cd -

if [ -d /usr/lib/systemd/system/ ]; then
  unit_dir=/usr/lib/systemd/system
else
  unit_dir=/etc/systemd/system
fi

install -Dm644 systemd/$service.service $unit_dir/$service.service
install -Dm644 conf.d/$service /etc/conf.d/$service
systemctl daemon-reload


echo edit /etc/prometheus/prometheus_machinist_gateway.yml
echo systemctl enable $service
echo systemctl start $service

