#!/usr/bin/env python3
"""
Extracts all attributes for the result object of schema files and displays
them nicely. Meant as a helper for documentation.
"""

import json
#import jsonschema
import jsonref
import os
import sys

def get_attributes(attribute: dict):
    '''
    Gets JSON attribute (dict) and returns description.
    If attribute is a JSON object or JSON array, it calls
    itself recursively.
    '''
    for attrib, value in attribute.items():
        print(f'''"{attrib}": {value['description']}''')

        # If "oneOf" has been found, we have to go through each entry.
        if 'oneOf' in value:
            for entry in value['oneOf']:
                #if entry['type'] 
                #get_attributes(entry)
                pass
            continue

        # If attribute is an object or array, it can contain sub-attributes.
        if value['type'] == 'object':
            get_attributes(value['properties'])
        if value['type'] == 'array':
            get_attributes(value['items']['properties'])


def main():

    # Walk through parameters (schema files).
    for file in sys.argv:

        # Loading file content as JSON.
        try:
            with open(file, 'r') as f:
                content = jsonref.JsonRef.replace_refs(json.load(f))
        except Exception as err:
            print(err, file=sys.stderr)
            continue

        # Extract attributes of schema.
        get_attributes(content['properties'])
        


if __name__ == '__main__':
    main()


