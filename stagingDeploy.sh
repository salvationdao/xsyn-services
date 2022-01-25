#!/bin/bash
# build binary
env GOOS=linux GOARCH=amd64 go build -o ./cmd/platform/passport-server ./cmd/platform/
# upload binary
scp ./cmd/platform/passport-server "root@sale.supremacy.fi:/root/passport-server/passport-server"
#scp passport-server-staging.env "root@sale.supremacy.fi:/home/passport-server/passport-server-staging.env"
# upload nginx config
scp ./passport-server.conf "root@sale.supremacy.fi:/home/passport-server/passport-server.conf"
# upload assets
scp -r ./asset "root@sale.supremacy.fi:/home/passport-server/asset"
# upload migrations
scp -r ./db/migrations "root@sale.supremacy.fi:/home/passport-server/migrations"
# upload migrate binary
scp ./bin/migrate "root@sale.supremacy.fi:/home/passport-server/migrate"
# run migrations
ssh root@sale.supremacy.fi 'cd /home/passport-server/ && source ./passport-server-staging.env && ./migrate -database "postgres://$LOCAL_DEV_DB_USER:$LOCAL_DEV_DB_PASS@$LOCAL_DEV_DB_HOST:$LOCAL_DEV_DB_PORT/$LOCAL_DEV_DB_DATABASE?sslmode=disable" -path ./migrations drop -f'
# run migrations
ssh root@sale.supremacy.fi 'cd /home/passport-server/ && source ./passport-server-staging.env && ./migrate -database "postgres://$LOCAL_DEV_DB_USER:$LOCAL_DEV_DB_PASS@$LOCAL_DEV_DB_HOST:$LOCAL_DEV_DB_PORT/$LOCAL_DEV_DB_DATABASE?sslmode=disable" -path ./migrations up'
# run seed
ssh root@sale.supremacy.fi 'cd /home/passport-server/ && ./passport-server db'
# move binary and restart services
ssh root@sale.supremacy.fi 'cd passport-server/ && chown passport-server:passport-server /root/passport-server/passport-server;systemctl stop passport-server;mv /root/passport-server/passport-server /home/passport-server/passport-server;systemctl start passport-server && sudo systemctl restart nginx'
