#!/usr/bin/env bash
#
# gh-dl-release! It works!
# 
# Adapted from
#  https://gist.github.com/maxim/6e15aa45ba010ab030c4
# 
# This script downloads an asset from latest or specific Github release of a
# private repo. Feel free to extract more of the variables into command line
# parameters.
#
# PREREQUISITES
#
# curl, jq
#
# USAGE
#
# Set all the variables inside the script, make sure you chmod +x it, then
# to download specific version:
#
#     gh-dl-release 2.1.1
#
# to download latest version:
#
#     gh-dl-release latest
#
# If your version/tag doesn't match, the script will exit with error.
source /home/ubuntu/.profile # load GITHUB_PAT
TOKEN=$GITHUB_PAT
REPO="ninja-syndicate/passport-server"
VERSION=$1                       # tag name or the word "latest"
FILE_PREFIX="passport-api"
GITHUB="https://api.github.com"

function release_curl() {
  curl -H "Authorization: token $TOKEN" \
       -H "Accept: application/vnd.github.v3.raw" \
       $@
}

if [ -z "$VERSION" ]; then
  echo " USAGE" >&2
  echo "" >&2
  echo " Set environment var \$GITHUB_PAT" >&2
  echo "" >&2
  echo "" >&2
  echo " to download specific version:" >&2
  echo "" >&2
  echo "     gh-dl-release 2.1.1" >&2
  echo "" >&2
  echo " to download latest version:" >&2
  echo "" >&2
  echo "     gh-dl-release latest" >&2
  echo "" >&2
  echo " If your version/tag doesn't match, the script will exit with error." >&2
  exit 1
fi;

if [ "$VERSION" = "latest" ]; then
  # Github should return the latest release first.
  parser_id=".assets | map(select(.name | contains(\"$FILE_PREFIX\")))[0].id"
  asset_id=`release_curl -s $GITHUB/repos/$REPO/releases/latest | jq "$parser_id"`
  parser_name=".assets | map(select(.name | contains(\"$FILE_PREFIX\")))[0].name"
  asset_name=`release_curl -s $GITHUB/repos/$REPO/releases/latest | jq --raw-output "$parser_name"`
else
  parser_id=". | map(select(.tag_name == \"$VERSION\"))[0].assets | map(select(.name | contains(\"$FILE_PREFIX\")))[0].id"
  asset_id=`release_curl -s $GITHUB/repos/$REPO/releases | jq "$parser_id"`
  parser_name=". | map(select(.tag_name == \"$VERSION\"))[0].assets | map(select(.name | contains(\"$FILE_PREFIX\")))[0].name"
  asset_name=`release_curl -s $GITHUB/repos/$REPO/releases | jq --raw-output "$parser_name"`
fi;

if [ -z "$asset_id" ]; then
  echo "ERROR: version or asset not found for $VERSION" >&2
  exit 1
fi;

if [ -t 1 ] ; then echo "Getting $asset_name"; fi # echo only if terminal

curl -H "Authorization: token $TOKEN" \
     -H "Accept:application/octet-stream" \
     --location --remote-header-name -s -o $asset_name \
     https://api.github.com/repos/$REPO/releases/assets/$asset_id \

echo "downloaded $(pwd)/$asset_name"

#
# USAGE
#
# Set all the variables inside the script, make sure you chmod +x it
#
#       change_version.sh <version>
#
# If your version/tag doesn't match, the script will exit with error.

CLIENT="ninja_syndicate"
PACKAGE="passport-api"
TARGET="$(pwd)/${PACKAGE}_$1"

cd /usr/share/$CLIENT

if [ -z "$1" ]; then
  echo "" >&2
  echo " USAGE" >&2
  echo "" >&2
  echo "" >&2
  echo " to bring an existing version online:" >&2
  echo "" >&2
  echo "     version_change target_dir" >&2
  echo "" >&2
  echo " If the directory doesn't exist and a tar does then it will be untared" >&2
  exit 1
fi;

if [ ! -d $TARGET ] ; then
  if [ -f $TARGET.tar.gz ];
    then tar -xvf $TARGET.tar.gz;
    else
      echo "Nither '$TARGET' or '$TARGET.tar.gz' was found in '/usr/share/$CLIENT'" >&2
      exit 2
  fi;
fi

VER=$(grep -oP 'Version=\K[0-9]+' /usr/share/${CLIENT}/${PACKAGE}_online/BuildInfo.txt || echo "0")
YMDHMS=$(date +'%Y%m%d%H%M%S')
DBDIR="/usr/share/${CLIENT}/${PACKAGE}_online/db_copy"
mkdir -p $DBDIR
DBFILE="$DBDIR/$PACKAGE_$YMDHMS.sql"

# Start the change over

source ${PACKAGE}_online/init/${PACKAGE}.env
cp ${PACKAGE}_online/init/${PACKAGE}.env $TARGET/init/${PACKAGE}.env

source /home/ubuntu/.profile # load PGPASSWORD

# Cant use the project default user due to adjusted permisions on some tables

systemctl stop ${PACKAGE}
$TARGET/migrate -database "postgres://${PASSPORT_DATABASE_USER}:${PASSPORT_DATABASE_PASS}@${PASSPORT_DATABASE_HOST}:${PASSPORT_DATABASE_PORT}/${PASSPORT_DATABASE_NAME}" -path $TARGET/migrations up

ln -Tfsv $TARGET $(pwd)/${PACKAGE}_online

# Ensure ownership
chown -R ${PACKAGE}:${PACKAGE} .

systemctl daemon-reload
systemctl restart ${PACKAGE}
nginx -t && nginx -s reload

echo "passport ${VERSION} ready!"

exit 0

