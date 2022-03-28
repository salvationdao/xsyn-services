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

CLIENT="ninja_syndicate"
PACKAGE="passport-api"
TARGET="$(pwd)/${PACKAGE}_$1"


if [ ! -d $TARGET ] ; then
  if [ -f $TARGET.tar.gz ];
    then tar -xvf $TARGET.tar.gz;
    else
      echo "Nither '$TARGET' or '$TARGET.tar.gz' was found in '/usr/share/$CLIENT'" >&2
      exit 2
  fi;
fi

YMDHMS=$(date +'%Y%m%d%H%M%S')
DBDIR="/usr/share/${CLIENT}/${PACKAGE}_online/db_copy"
mkdir -p $DBDIR
DBFILE="$DBDIR/$PACKAGE_$YMDHMS.sql"

# Start the change over

# Cant use the project default user due to adjusted permisions on some tables
pg_dump --dbname="$PASSPORT_DATABASE_NAME" --host="$PASSPORT_DATABASE_HOST" --port="$PASSPORT_DATABASE_PORT" --username="postgres" > ${DBFILE}

if [ ! -s "${DBFILE}" ]; then
    echo "db copy is zero size" >&2
    exit 3
fi

ls -lh "${DBFILE}"
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
