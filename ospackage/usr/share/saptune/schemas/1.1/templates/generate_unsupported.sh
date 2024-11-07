#!/usr/bin/bash
set -u

# Builds schema definitions in ./ for commands without JSON support.
# Existing files get overwritten if FORCE=2!
#
# Usage: FORCE=1 ./generate_unsupported.sh

UNSUPPORTED_COMMANDS=( 
    "saptune check"
    "saptune daemon start" 
    "saptune daemon stop"
    "saptune service start"
    "saptune service reload"
    "saptune service restart"
    "saptune service stop"
    "saptune service enable"
    "saptune service disable"
    "saptune service enablestart"
    "saptune service disablestop"
    "saptune service apply"
    "saptune service revert"
    "saptune service takeover"
    "saptune note apply"
    "saptune note simulate"
    "saptune note customise"
    "saptune note customize"
    "saptune note create"
    "saptune note edit"
    "saptune note show"
    "saptune note revert"
    "saptune note delete"
    "saptune note rename"
    "saptune note refresh"
    "saptune note revertall"
    "saptune revert all"
    "saptune solution apply"
    "saptune solution change"
    "saptune solution simulate"
    "saptune solution revert"
    "saptune solution create"
    "saptune solution edit"
    "saptune solution delete"
    "saptune solution rename"
    "saptune solution show"
    "saptune staging status"
    "saptune staging is-enabled"
    "saptune staging enable"
    "saptune staging disable"
    "saptune staging list"
    "saptune staging diff"
    "saptune staging analysis"
    "saptune staging release"
    "saptune configure"
    "saptune refresh"
    "saptune log status"
    "saptune log set"
    "saptune lock remove"
    "saptune help"
)

function write_template() {
    local cmd="${1}"
    local file="${2}" 
    cat > "${file}" <<-EOF
{% extends "common.schema.json.template" %}

{% block command %}${cmd}{% endblock %}

{% block description %}Describes the output of '{{ self.command() }}.{% endblock %}

{% block result_required %}["implemented"]{% endblock %}

{% block result_properties %}
                "implemented": {
                    "description": "Indicates that JSON output has not yet been implemented yet.",
                    "type": "boolean",
                    "enum": [false]
                }            
{% endblock %}
EOF
}


# Exit if FORCE isn't set.
[ "${FORCE:=0}" == '0' ] && { echo 'Variable FORCE not set to 1 or 2, so we terminate.' >&2 ; exit 1 ; }

# Walk through the commands and generate JSON schema.
for ((c=0; c<${#UNSUPPORTED_COMMANDS[*]}; c++)) ; do

    cmd="${UNSUPPORTED_COMMANDS[c]}"
    schema_file="${cmd// /_}.schema.json.template"

    if [ -e "${schema_file}" ] ; then 
        if [ "${FORCE:=0}" != '2' ] ; then 
            echo -e "[\033[33mWARN\033[39m] ${schema_file} Variable FORCE not set to 2, so existing file \"${schema_file}\" does not get overwritten."
            continue
        fi
    fi
    
    write_template "${cmd}" "${schema_file}"
    if [ $? -eq 0 ] ; then
        echo -e "[\033[32m OK \033[39m] ${schema_file} created."
    else
        echo -e "[\033[31mFAIL\033[39m] ${schema_file} could not be created."
    fi
done

# Bye.
exit 0 