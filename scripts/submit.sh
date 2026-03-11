#!/usr/bin/env bash
# submit.sh: send form data from a JSON file to the resume server.
# Usage: ./scripts/submit.sh <endpoint> <jsonfile> <output>

endpoint="$1"
json="$2"
output="$3"

if [ -z "$endpoint" ] || [ -z "$json" ] || [ -z "$output" ]; then
    echo "Usage: $0 endpoint jsonfile output" >&2
    exit 1
fi

# build curl -F arguments from JSON keys/values
curl_args=()
while IFS= read -r line; do
    # line contains key=value
    key=${line%%=*}
    val=${line#*=}
    if [[ $val == @* ]]; then
        # file upload
        curl_args+=("-F" "$key=$val")
    elif [[ -f $val ]]; then
        # treat as file if path exists
        curl_args+=("-F" "$key=@$val")
    else
        curl_args+=("-F" "$key=$val")
    fi
done < <(jq -r 'to_entries | .[] | "\(.key)=\(.value)"' "$json")

curl -s -X POST "${curl_args[@]}" "$endpoint" -o "$output"
