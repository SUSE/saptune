#!/bin/bash

for source in snippets/*.html ; do
    dest="pages/${source##*/}"
    echo "Genertaing ${dest}"
    jinja "${source}" > "${dest}"
done

exit 0