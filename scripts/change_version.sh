


ln -Tfsv $TARGET $(pwd)/${PACKAGE}_online

# Ensure ownership
chown -R ${PACKAGE}:${PACKAGE} .

