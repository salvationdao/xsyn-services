#!/usr/bin/env bash
#
# USAGE
#
# Set all the variables inside the script, make sure you chmod +x it
#
#       change_version.sh <version>
#
# If your version/tag doesn't match, the script will exit with error.

set -e

# Start the change over

# Cant use the project default user due to adjusted permisions on some tables

echo "Proceed with migrations? (y/N)"
read PROCEED
if [[ $PROCEED != "y" ]]; then exit 4; fi

systemctl stop ${PACKAGE}
$TARGET/migrate -database "postgres://${PASSPORT_DATABASE_USER}:${PASSPORT_DATABASE_PASS}@${PASSPORT_DATABASE_HOST}:${PASSPORT_DATABASE_PORT}/${PASSPORT_DATABASE_NAME}" -path $TARGET/migrations up

ln -Tfsv $TARGET $(pwd)/${PACKAGE}_online

# Ensure ownership
chown -R ${PACKAGE}:${PACKAGE} .

systemctl daemon-reload
systemctl restart ${PACKAGE}
nginx -t && nginx -s reload
